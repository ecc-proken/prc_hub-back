package flags

import (
	"flag"
)

type Flags struct {
	Port               *uint
	LogLevel           *uint
	GzipLevel          *uint
	MysqlHost          *string
	MysqlPORT          *uint
	MysqlDB            *string
	MysqlUser          *string
	MysqlPasswd        *string
	JwtIssuer          *string
	JwtSecret          *string
	GithubClientId     *string
	GithubClientSecret *string
	AdminEmail         *string
	AdminPasswd        *string
}

var flags Flags

func Get() Flags {
	if flags.Port == nil {
		return parse()
	}
	return flags
}

// 優先度: command line params > env variables > default value
func parse() Flags {
	flags = Flags{
		flag.Uint("port", getUintEnv("PORT", 1323), "Server port"),
		flag.Uint("log-level", getUintEnv("LOG_LEVEL", 2), "Log level (1: 'DEBUG', 2: 'INFO', 3: 'WARN', 4: 'ERROR', 5: 'OFF', 6: 'PANIC', 7: 'FATAL'"),
		flag.Uint("gzip-level", getUintEnv("GZIP_LEVEL", 6), "Gzip compression level"),
		flag.String("mysql-host", getEnv("MYSQL_HOST", "db"), "MySQL host"),
		flag.Uint("mysql-port", getUintEnv("MYSQL_PORT", 3306), "MySQL port"),
		flag.String("mysql-database", getEnv("MYSQL_DATABASE", "prc_hub-api"), "MySQL database"),
		flag.String("mysql-user", getEnv("MYSQL_USER", "prc_hub-api"), "MySQL user"),
		flag.String("mysql-password", getEnv("MYSQL_PASSWORD", ""), "MySQL password"),
		flag.String("jwt-issuer", getEnv("JWT_ISSUER", "prc_hub-api"), "JWT issuer"),
		flag.String("jwt-secret", getEnv("JWT_SECRET", ""), "JWT secret"),
		flag.String("github-client-id", getEnv("GITHUB_CLIENT_ID", ""), "GitHub client id"),
		flag.String("github-client-secret", getEnv("GITHUB_CLIENT_SECRET", ""), "GitHub client secret"),
		flag.String("admin-email", getEnv("ADMIN_EMAIL", ""), "Admin user email"),
		flag.String("admin-password", getEnv("ADMIN_PASSWD", ""), "Admin user password"),
	}

	flag.Parse()
	return flags
}
