package ratelimiter

import "time"

type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerFrame int
	TimeFrame       time.Duration
	Enabled         bool
}
