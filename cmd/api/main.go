package main

import (
	"context"
	"database/sql"
	"flag"
	"sync"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/mailer"
	"github.com/hashicorp/go-hclog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	_ "github.com/lib/pq"

	_ "github.com/Kaungmyatkyaw2/book-store-api/cmd/api/docs"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	jwt struct {
		secret string
	}
	googleOauth struct {
		redirectUrl  string
		clientID     string
		clientSecret string
	}
}

type application struct {
	config      config
	logger      hclog.Logger
	models      data.Models
	mailer      mailer.IMailer
	wg          sync.WaitGroup
	googleOauth *oauth2.Config
}

// @title Book Store API
// @version 1.0
// @description This is book store API built using Go and httprouter
// @host localhost:4000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {

	var cfg config

	logger := hclog.Default()

	flag.IntVar(&cfg.port, "port", 4000, "API server port.")
	flag.StringVar(&cfg.env, "env", "development", "environment")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "Postgresql DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "74644856878fe8", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "2eda21f9e2176d", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.eillion.net>", "SMTP sender")

	flag.StringVar(&cfg.jwt.secret, "jwt-secret", "", "JWT secret")

	flag.StringVar(&cfg.googleOauth.redirectUrl, "oauth-redirect-url", "http://localhost:4000/v1/auth/google/callback", "Google oauth redirect url")
	flag.StringVar(&cfg.googleOauth.clientID, "oauth-client-id", "", "Google oauth client id")
	flag.StringVar(&cfg.googleOauth.clientSecret, "oauth-client-secret", "", "Google oauth client secret")

	flag.Parse()

	db, err := openDB(cfg)

	if err != nil {
		logger.Error(err.Error())
	}

	defer db.Close()

	logger.Info("database connection pool established!")

	googleOauthConfig := oauth2.Config{
		ClientID:     cfg.googleOauth.clientID,
		ClientSecret: cfg.googleOauth.clientSecret,
		RedirectURL:  cfg.googleOauth.redirectUrl,
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}

	app := &application{
		config:      cfg,
		logger:      logger,
		models:      data.NewModels(db),
		mailer:      mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		googleOauth: &googleOauthConfig,
	}

	err = app.serve()

	if err != nil {
		logger.Error(err.Error())
	}

}

func openDB(cfg config) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
