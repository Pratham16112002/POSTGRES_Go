package main

import (
	"Blog/internal/db"
	"Blog/internal/env"
	"Blog/internal/mailer"
	"Blog/internal/store"
	"time"

	"go.uber.org/zap"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":3001"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:1234@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:         env.GetString("ENV", "development"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:3002"),
		mail: mailConfig{
			exp:       time.Minute * 10,
			fromEmail: env.GetString("FROM_EMAIL", "blogspot@support.com"),
			sendGrid: SendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
	}
	// Mailer
	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	// Database
	db, err := db.NewDB(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime, logger)
	if err != nil {
		logger.Fatal("Data base connection failed", err)
	}
	defer db.Close()
	store := store.NewPostgresStore(db)
	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}
	logger.Info("Server is starting on %v\n", cfg.addr)
	mux := app.mount()
	logger.Fatal(app.run(mux))
}
