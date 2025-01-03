package main

import (
	ratelimiter "Blog/internal/rateLimiter"
	"Blog/internal/store"
	"testing"

	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	mockStore := store.NewMockStore()
	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(cfg.rateLimiter.RequestPerFrame, cfg.rateLimiter.TimeFrame)
	return &application{
		logger:      logger,
		store:       mockStore,
		config:      cfg,
		rateLimiter: rateLimiter,
	}
}

// func executeRequest(req *http.Request, mux *chi.Mux) {
// 	rr := httptest.NewRecorder()
// 	mux.ServeHTTP(rr, req)
// }

// func checkResponse(t *testing.T, expected, actual int) {
// 	if expected != actual {
// 		t.Errorf("Expected response code %d, Got %d", expected, actual)
// 	}
// }
