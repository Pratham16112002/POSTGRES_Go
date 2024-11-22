package db

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

func NewDB(addr string, maxOpenConns int, maxIdleConns int, maxIdleTime string, logger *zap.SugaredLogger) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	parsedTime, _ := time.ParseDuration(maxIdleTime)
	db.SetConnMaxIdleTime(parsedTime)
	ctx, cnl_fn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cnl_fn()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	logger.Infof("database connection successfull", db.Stats())
	return db, err
}
