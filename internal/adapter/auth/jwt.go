package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"api-gateway/internal/domain/auth"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrTokenExpired  = errors.New("token expired")
	ErrInvalidIssuer = errors.New("invalid issuer")
)

type JWTValidator struct {
	secret []byte
	issuer string
}

func NewJWTValidator(secret, issuer string) *JWTValidator {
	return &JWTValidator{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (v *JWTValidator) Validate(ctx context.Context, tokenString string) (*auth.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return v.secret, nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if v.issuer != "" {
		iss, ok := claims["iss"].(string)
		if !ok || iss != v.issuer {
			return nil, ErrInvalidIssuer
		}
	}

	subject, _ := claims["sub"].(string)
	name, _ := claims["name"].(string)
	admin, _ := claims["admin"].(bool)

	var expiresAt, issuedAt int64
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = int64(exp)
	}
	if iat, ok := claims["iat"].(float64); ok {
		issuedAt = int64(iat)
	}

	raw := make(map[string]interface{})
	for k, val := range claims {
		raw[k] = val
	}

	return &auth.Claims{
		Subject:   subject,
		Name:      name,
		Admin:     admin,
		Issuer:    v.issuer,
		ExpiresAt: expiresAt,
		IssuedAt:  issuedAt,
		Raw:       raw,
	}, nil
}

func GenerateToken(secret, issuer, subject, name string, admin bool, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   subject,
		"name":  name,
		"admin": admin,
		"iss":   issuer,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(duration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
