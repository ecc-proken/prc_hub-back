package main

import (
	"fmt"
	"prc_hub-api/flags"
	handler_events "prc_hub-api/handler/events"
	handler_oauth2 "prc_hub-api/handler/oauth_provider"
	handler_users "prc_hub-api/handler/users"
	"prc_hub-api/jwt"
	"prc_hub-api/migration"
	"prc_hub-api/mysql"
	"prc_hub-api/oauth2/github"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return err
	}
	return nil
}

func main() {
	// コマンドライン引数 / 環境変数 の取得
	f := flags.Get()

	// echoサーバーのインスタンス生成
	e := echo.New()
	// ログレベルの設定
	e.Logger.SetLevel(log.Lvl(*f.LogLevel))
	// Gzipの圧縮レベル設定
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: int(*f.GzipLevel),
	}))
	// CORSの設定
	if f.AllowOrigins != nil && len(f.AllowOrigins) != 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: f.AllowOrigins,
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		}))
	}
	// structの変数を検証するvalidatorをechoに設定
	e.Validator = &CustomValidator{validator: validator.New()}

	// mysqlに接続するクライアントの設定
	e.Logger.Info(mysql.SetDSNTCP(*f.MysqlUser, *f.MysqlPasswd, *f.MysqlHost, int(*f.MysqlPORT), *f.MysqlDB))

	// Adminユーザーのマイグレーション
	adminFound, updated, invalidEmail, usedEmail, err := migration.MigrateAdminUser(*f.AdminEmail, *f.AdminPasswd)
	if err != nil {
		e.Logger.Fatal(err.Error())
		return
	}
	if invalidEmail || usedEmail {
		// Adminユーザーのemailが使用済みまたは不正
		e.Logger.Fatalf("Admin email already used or invalid. %s", *f.AdminEmail)
	} else if !adminFound || updated {
		// Adminユーザーの追加成功
		e.Logger.Info("Migrate admin user successful.")
	}

	if *f.GithubClientId != "" && *f.GithubClientSecret != "" {
		err = github.SetClient(*f.GithubClientId, *f.GithubClientSecret)
		if err != nil {
			e.Logger.Error(err.Error())
		}
	} else {
		e.Logger.Info("ClientId and ClientSecret not found, GitHub OAuth2 disabled")
	}

	// JWTの設定
	e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     &jwt.JwtCustumClaims{},
		SigningKey: []byte(*f.JwtSecret),
		Skipper: func(c echo.Context) bool {
			// 公開エンドポイントのJWT認証をスキップ
			return c.Path() == "/users" && c.Request().Method == "POST" ||
				c.Path() == "/users/:provider/register" && c.Request().Method == "POST" ||
				c.Path() == "/users/sign_in" && c.Request().Method == "POST" ||
				c.Path() == "/events" && c.Request().Method == "GET" ||
				c.Path() == "/events/:id" && c.Request().Method == "GET"
		},
	}))

	// 公開エンドポイント
	e.POST("/users", handler_users.Post)
	e.POST("/users/oauth2/:provider/register", handler_oauth2.Register)
	e.POST("/users/sign_in", handler_users.SignIn)
	e.GET("/events", handler_events.Get)
	e.GET("/events/:id", handler_events.GetById)

	// JWT認証必須エンドポイント
	e.GET("/users", handler_users.Get)
	e.PATCH("/users", handler_users.PatchOwn)
	e.DELETE("/users", handler_users.DeleteOwn)
	e.GET("/users/:id", handler_users.GetById)
	e.PATCH("/users/:id", handler_users.PatchById)
	e.DELETE("/users/:id", handler_users.DeleteById)
	e.POST("/users/oauth2/:provider", handler_oauth2.Post)
	e.DELETE("/users/oauth2/:provider", handler_oauth2.Delete)
	e.POST("/events", handler_events.Post)
	e.PATCH("/events/:id", handler_events.PatchById)
	e.DELETE("/events/:id", handler_events.DeleteById)
	e.POST("/events/:id/:dt_id/participate", handler_events.Participate)
	e.DELETE("/events/:id/:dt_id/participate", handler_events.Unparticipate)
	e.POST("/events/:id/webhook/line_notify", handler_events.LineNotify)
	e.POST("/events/:id/documents/webhook/line_notify", handler_events.LineNotifyDocuments)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *f.Port)))
}
