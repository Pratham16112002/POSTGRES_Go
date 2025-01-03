package main

import (
	"Blog/internal/auth"
	"Blog/internal/env"
	"Blog/internal/mailer"
	ratelimiter "Blog/internal/rateLimiter"
	"Blog/internal/store"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/swaggo/swag/example/override/docs"
	"go.uber.org/zap"
)

type application struct {
	config      config
	store       store.Storage
	logger      *zap.SugaredLogger
	mailer      mailer.Client
	auth        auth.JWTAuthenticator
	rateLimiter ratelimiter.Limiter
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	frontendURL string
	mail        mailConfig
	auth        authConfig
	rateLimiter ratelimiter.Config
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	user string
	pass string
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	apiKey    string
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{env.GetString("LOCAL_FRONTEND_URL", "http://localhost:3000")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	if app.config.rateLimiter.Enabled {
		r.Use(app.RateLimiterMiddleware)
	}
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.With(app.BasicAuthMiddleware())
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthenTokenMiddleware())
			r.Post("/", app.createPostHandler)
			r.Route("/{postId}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)
				r.Get("/", app.getPostHanlder)
				r.Delete("/", app.CheckPostOwnership("admin", app.deletePostHandler))
				r.Patch("/", app.CheckPostOwnership("moderator", app.updatePostHandler))
				r.Post("/comments", app.postCommentHandler)
			})
		})
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.userActivationHandler)
			r.Route("/{userId}", func(r chi.Router) {
				r.Use(app.AuthenTokenMiddleware())
				r.Use(app.usersContextMiddleware)
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.AuthenTokenMiddleware())
				r.Get("/feed", app.getUserFeedHandler)
			})
		})
		// Public routes
		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.userRegisterHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})
	// posts
	// users
	// auth
	return r
}

func (app *application) run(mux *chi.Mux) error {

	docs.SwaggerInfo.Version = version

	srv := http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		app.logger.Infow("signal caught", "sginal", s.String())
		shutdown <- srv.Shutdown(ctx)
	}()
	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdown
	if err != nil {
		return err
	}
	app.logger.Infow("server has stopped", "addr", app.config.addr)
	return nil
}
