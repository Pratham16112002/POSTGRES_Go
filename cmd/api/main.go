package main

import (
	"Blog/internal/env"
	"Blog/internal/store"
	"log"
)

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":3001"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:adminpasword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	store := store.NewPostgresStore(nil)
	app := &application{
		config: cfg,
		store:  store,
	}
	log.Printf("Server is starting on %v\n", cfg.addr)
	mux := app.mount()
	log.Fatal(app.run(mux))
}
