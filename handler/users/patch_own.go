package users

import (
	"encoding/json"
	"net/http"
	"prc_hub-api/flags"
	"prc_hub-api/jwt"
	"prc_hub-api/users"

	jwtGo "github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func PatchOwn(c echo.Context) (err error) {
	// jwtトークン確認
	t := c.Get("user").(*jwtGo.Token)
	claims, err := jwt.CheckToken(*flags.Get().JwtIssuer, t)
	if err != nil {
		c.Logger().Debug("401: " + err.Error())
		return c.JSONPretty(http.StatusNotFound, map[string]string{"message": err.Error()}, "	")
	}
	user_id := claims.Id

	// リクエストボディをバインド
	p := new(users.PatchBody)
	if err = c.Bind(p); err != nil {
		// 400: Bad request
		c.Logger().Debug("400: " + err.Error())
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": err.Error()}, "	")
	}

	// リクエストボディを検証
	if err = c.Validate(p); err != nil {
		// 422: Unprocessable entity
		c.Logger().Debug("422: " + err.Error())
		return c.JSONPretty(http.StatusUnprocessableEntity, map[string]string{"message": err.Error()}, "	")
	}
	if err = p.Validate(); err != nil {
		// 400: Bad request
		c.Logger().Debug("400: " + err.Error())
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": err.Error()}, "	")
	}
	if !claims.Admin {
		if p.PostEventAvailabled != nil || p.Admin != nil {
			// Admin権限がない場合、変更不可
			c.Logger().Debug("403: you cannot change user authority")
			return c.JSONPretty(http.StatusForbidden, map[string]string{"message": "you cannot change user authority"}, "	")
		}
	} else if claims.MigratedAdmin &&
		(p.PostEventAvailabled != nil && !*p.PostEventAvailabled || p.Admin != nil && !*p.Admin) {
		// Adminの権限は変更不可
		c.Logger().Debug("400: cannot change admin user authority")
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": "cannot change admin user authority"}, "	")
	}

	// 更新
	u, invalidEmail, usedEmail, notFound, err := users.Patch(user_id, *p)
	if err != nil {
		c.Logger().Fatal(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}
	if invalidEmail {
		// 422: Unprocessable entity
		c.Logger().Debug("422: invalid email")
		return c.JSONPretty(http.StatusUnprocessableEntity, map[string]string{"message": "invalid email"}, "	")
	}
	if usedEmail {
		// 400: Bad request
		c.Logger().Debug("400: email already used")
		return c.JSONPretty(http.StatusBadRequest, map[string]string{"message": "email already used"}, "	")
	}
	if notFound {
		// 404: Not found
		c.Logger().Debug("404: user not found")
		return c.JSONPretty(http.StatusNotFound, map[string]string{"message": "user not found"}, "	")
	}

	// トークンを生成
	token, err := jwt.GenerateToken(u, *flags.Get().JwtIssuer, *flags.Get().JwtSecret)
	if err != nil {
		c.Logger().Fatal(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}

	// jsonにjwtトークンを追加
	b, err := json.Marshal(u)
	if err != nil {
		c.Logger().Fatal(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}
	m := map[string]interface{}{"token": token}
	err = json.Unmarshal(b, &m)
	if err != nil {
		c.Logger().Fatal(err)
		return c.JSONPretty(http.StatusInternalServerError, map[string]string{"message": err.Error()}, "	")
	}

	// 200: Success
	c.Logger().Debug("200: patch user successful")
	return c.JSONPretty(http.StatusOK, m, "	")
}
