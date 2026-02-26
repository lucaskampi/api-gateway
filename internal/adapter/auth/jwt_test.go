package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestJWTValidator_Validate_ValidToken(t *testing.T) {
	secret := "test-secret"
	issuer := "test-issuer"

	token, err := GenerateToken(secret, issuer, "123", "Test User", true, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	validator := NewJWTValidator(secret, issuer)
	claims, err := validator.Validate(context.Background(), token)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if claims.Subject != "123" {
		t.Errorf("expected subject 123, got %s", claims.Subject)
	}

	if claims.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", claims.Name)
	}

	if !claims.Admin {
		t.Error("expected admin to be true")
	}
}

func TestJWTValidator_Validate_InvalidToken(t *testing.T) {
	validator := NewJWTValidator("secret", "issuer")

	_, err := validator.Validate(context.Background(), "invalid.token.here")

	if err == nil {
		t.Error("expected error for invalid token")
	}

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTValidator_Validate_WrongSecret(t *testing.T) {
	validator := NewJWTValidator("secret", "issuer")

	token, _ := GenerateToken("wrong-secret", "issuer", "123", "Test", true, time.Hour)
	_, err := validator.Validate(context.Background(), token)

	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestJWTValidator_Validate_WrongIssuer(t *testing.T) {
	validator := NewJWTValidator("secret", "issuer")

	token, _ := GenerateToken("secret", "wrong-issuer", "123", "Test", true, time.Hour)
	_, err := validator.Validate(context.Background(), token)

	if err == nil {
		t.Error("expected error for wrong issuer")
	}

	if !errors.Is(err, ErrInvalidIssuer) {
		t.Errorf("expected ErrInvalidIssuer, got %v", err)
	}
}

func TestJWTValidator_Validate_ExpiredToken(t *testing.T) {
	validator := NewJWTValidator("secret", "issuer")

	token, _ := GenerateToken("secret", "issuer", "123", "Test", true, -time.Hour)
	_, err := validator.Validate(context.Background(), token)

	if err == nil {
		t.Error("expected error for expired token")
	}
}
