package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string
	aud    string
	iss    string
}

func NewJWT(secret, aud, iss string) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret: secret,
		aud:    aud,
		iss:    iss,
	}
}

func (j *JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secret))

	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func (j *JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		// Checking weather the signing method was the same when it was generated
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected singning method %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	}, jwt.WithAudience(j.aud), jwt.WithExpirationRequired(), jwt.WithIssuer(j.iss), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
}
