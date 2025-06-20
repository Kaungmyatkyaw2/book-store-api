package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/hashicorp/go-hclog"
)

func loadConfig(cfg *config) {

	flag.IntVar(&cfg.port, "port", getIntEnv("PORT", 4000), "API server port.")
	flag.StringVar(&cfg.env, "env", os.Getenv("ENV"), "environment")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DB_DSN"), "Postgresql DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", getIntEnv("SMTP_PORT", 25), "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("Greenlight <no-reply@greenlight.eillion.net>"), "SMTP sender")

	flag.StringVar(&cfg.jwt.secret, "jwt-secret", os.Getenv("JWT_SECRET"), "JWT secret")

	flag.StringVar(&cfg.googleOauth.redirectUrl, "oauth-redirect-url", os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"), "Google oauth redirect url")
	flag.StringVar(&cfg.googleOauth.clientID, "oauth-client-id", os.Getenv("GOOGLE_OAUTH_CLIENT_ID"), "Google oauth client id")
	flag.StringVar(&cfg.googleOauth.clientSecret, "oauth-client-secret", os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"), "Google oauth client secret")

	flag.Parse()
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		hclog.Default().Error("Invalid environment alue detected: ", valueStr)
		return defaultValue
	}

	return value
}
