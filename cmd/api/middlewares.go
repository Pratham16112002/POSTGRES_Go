package main

import (
	"Blog/internal/store"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type UserCtxKey string

const authUser UserCtxKey = "user"

var (
	ErrUnAuthorized = errors.New("unauthorized")
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			// read the auth header
			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				app.basicAuthorizationError(res, req, errors.New("user not authorized"))
				return
			}
			// decode it
			authHeader_parts := strings.Split(authHeader, " ")
			if len(authHeader_parts) != 2 || authHeader_parts[0] != "Basic" {
				app.basicAuthorizationError(res, req, errors.New("auth token not found"))
				return
			}
			app.logger.Infow("msg", authHeader_parts)
			decoded, err := base64.StdEncoding.DecodeString(authHeader_parts[1])
			if err != nil {
				app.basicAuthorizationError(res, req, err)
				return
			}
			creds := strings.SplitN(string(decoded), ":", 2)
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass
			if len(creds) != 2 && creds[0] != username && creds[1] != pass {
				app.basicAuthorizationError(res, req, errors.New("access denied"))
				return
			}
			next.ServeHTTP(res, req)
		})
	}

}

func (app *application) AuthenTokenMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			auth_header := req.Header.Get("Authorization")

			if auth_header == "" {
				app.authorizationError(res, req, errors.New("authorization header missing"))
				return
			}

			parts := strings.Split(auth_header, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				app.authorizationError(res, req, errors.New("authorization header is malformed"))
				return
			}

			token, err := app.auth.ValidateToken(parts[1])
			if err != nil {
				app.authorizationError(res, req, err)
				return
			}

			claims, _ := token.Claims.(jwt.MapClaims)

			userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

			if err != nil {
				app.authorizationError(res, req, err)
				return
			}
			ctx := req.Context()
			user, err := app.store.Users.GetUserById(ctx, userId)

			if err != nil {
				switch err {
				case store.ErrNotFound:
					app.notFoundError(res, req, err)
				default:
					app.internalServerError(res, req, err)
				}
				return
			}
			ctx = context.WithValue(ctx, authUser, user)
			next.ServeHTTP(res, req.WithContext(ctx))
		})
	}
}

func (app *application) CheckPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		user := getAuthUser(req)
		post := getPostFromCtx(req)

		if post.UserId == user.ID {
			next.ServeHTTP(res, req)
			return
		}
		ctx := req.Context()
		allowed, err := app.checkRolePrecedence(ctx, user, requiredRole)
		if !allowed {
			app.forbiddenError(res, req, ErrUnAuthorized)
			return
		}
		if err != nil {
			app.internalServerError(res, req, err)
			return
		}

		next.ServeHTTP(res, req)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetRoleByName(ctx, roleName)
	if err != nil {
		return false, store.ErrNotFound
	}
	if user.Role.Level < role.Level {
		return false, nil
	}
	return true, nil
}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(req.RemoteAddr); !allow {
				app.rateLimitExceededResponse(res, req, retryAfter.String())
				return
			}
		}
		next.ServeHTTP(res, req)
	})
}
