package main

import (
	"Blog/internal/auth"
	"Blog/internal/db"
	"Blog/internal/env"
	"Blog/internal/mailer"
	"Blog/internal/store"
	"time"

	"go.uber.org/zap"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":3002"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:1234@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env:         env.GetString("ENV", "development"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:3001/"),
		mail: mailConfig{
			exp:       time.Minute * 10,
			fromEmail: env.GetString("FROM_EMAIL", "pratham.209302249@muj.manipal.edu"),
			apiKey:    env.GetString("EMAIL_API_KEY", "3f8f43d7-15da-490b-8a2c-71bc7fe7506f"),
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "1234"),
			},
			token: tokenConfig{
				secret: env.GetString("JWT_SECRET", "12345"),
				exp:    time.Hour * 24 * 3,
				iss:    env.GetString("JWT_ISS", "jwthost"),
			},
		},
	}
	// JWT
	jwtAuth := auth.NewJWT(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)
	// Mailer
	mailer := mailer.NewMailer(cfg.mail.apiKey, cfg.mail.fromEmail)
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
		auth:   *jwtAuth,
	}
	logger.Info("Server is starting on %v\n", cfg.addr)
	mux := app.mount()
	logger.Fatal(app.run(mux))
}
