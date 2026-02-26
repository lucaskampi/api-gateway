package auth

import (
	"context"
)

type Claims struct {
	Subject   string
	Name      string
	Admin     bool
	Issuer    string
	ExpiresAt int64
	IssuedAt  int64
	Raw       map[string]interface{}
}

type TokenValidator interface {
	Validate(ctx context.Context, token string) (*Claims, error)
}
