package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func BenchmarkJWT_Validate(b *testing.B) {
	secret := "test-secret-key-for-benchmarking"
	issuer := "test-issuer"

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "user123",
		"name":  "Test User",
		"admin": false,
		"iss":   issuer,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		b.Fatal(err)
	}

	validator := NewJWTValidator(secret, issuer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(context.Background(), token)
	}
}

func BenchmarkJWT_Generate(b *testing.B) {
	secret := "test-secret-key-for-benchmarking"
	issuer := "test-issuer"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateToken(secret, issuer, "user123", "Test User", false, time.Hour)
	}
}

func BenchmarkJWT_ValidateInvalid(b *testing.B) {
	secret := "test-secret-key-for-benchmarking"
	issuer := "test-issuer"

	validator := NewJWTValidator(secret, issuer)
	invalidToken := "invalid.token.here"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(context.Background(), invalidToken)
	}
}
