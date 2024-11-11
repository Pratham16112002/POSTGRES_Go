package db

import (
	"database/sql"
	"time"
)

func NewDB(addr string, maxOpenConns int, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	parsedTime, _ := time.ParseDuration(maxIdleTime)
	db.SetConnMaxIdleTime(parsedTime)
	if err != nil {
		return nil, err
	}
	return db, err
}
