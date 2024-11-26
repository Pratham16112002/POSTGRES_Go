package auth

import "github.com/golang-jwt/jwt/v5"

type Authenticator interface {
	// accept the claims about the subject and return the token or error.
	GenerateToken(claims jwt.Claims) (string, error)
	// validate the token and return jwt
	ValidateToken(token string) (*jwt.Token, error)
}
