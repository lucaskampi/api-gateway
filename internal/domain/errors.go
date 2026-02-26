package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrCodeNotFound            ErrorCode = "ERR_NOT_FOUND"
	ErrCodeUnauthorized        ErrorCode = "ERR_UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "ERR_FORBIDDEN"
	ErrCodeUpstreamUnavailable ErrorCode = "ERR_UPSTREAM_UNAVAILABLE"
	ErrCodeRateLimited         ErrorCode = "ERR_RATE_LIMITED"
	ErrCodeTimeout             ErrorCode = "ERR_TIMEOUT"
	ErrCodeConfigInvalid       ErrorCode = "ERR_CONFIG_INVALID"
	ErrCodeInvalidToken        ErrorCode = "ERR_INVALID_TOKEN"
	ErrCodeTokenExpired        ErrorCode = "ERR_TOKEN_EXPIRED"
	ErrCodeInvalidIssuer       ErrorCode = "ERR_INVALID_ISSUER"
	ErrCodeInternalError       ErrorCode = "ERR_INTERNAL_ERROR"
	ErrCodeBadGateway          ErrorCode = "ERR_BAD_GATEWAY"
	ErrCodeServiceUnavailable  ErrorCode = "ERR_SERVICE_UNAVAILABLE"
)

type GatewayError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *GatewayError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *GatewayError) Unwrap() error {
	return e.Err
}

func (e *GatewayError) Is(target error) bool {
	if t, ok := target.(*GatewayError); ok {
		return e.Code == t.Code
	}
	return false
}

func (e *GatewayError) With(err error) *GatewayError {
	return &GatewayError{
		Code:    e.Code,
		Message: e.Message,
		Err:     err,
	}
}

var (
	ErrNotFound            = &GatewayError{Code: ErrCodeNotFound, Message: "resource not found"}
	ErrUnauthorized        = &GatewayError{Code: ErrCodeUnauthorized, Message: "unauthorized"}
	ErrForbidden           = &GatewayError{Code: ErrCodeForbidden, Message: "forbidden"}
	ErrUpstreamUnavailable = &GatewayError{Code: ErrCodeUpstreamUnavailable, Message: "upstream service unavailable"}
	ErrRateLimited         = &GatewayError{Code: ErrCodeRateLimited, Message: "rate limit exceeded"}
	ErrTimeout             = &GatewayError{Code: ErrCodeTimeout, Message: "request timeout"}
	ErrConfigInvalid       = &GatewayError{Code: ErrCodeConfigInvalid, Message: "invalid configuration"}
	ErrInvalidToken        = &GatewayError{Code: ErrCodeInvalidToken, Message: "invalid token"}
	ErrTokenExpired        = &GatewayError{Code: ErrCodeTokenExpired, Message: "token expired"}
	ErrInvalidIssuer       = &GatewayError{Code: ErrCodeInvalidIssuer, Message: "invalid token issuer"}
	ErrInternalError       = &GatewayError{Code: ErrCodeInternalError, Message: "internal server error"}
	ErrBadGateway          = &GatewayError{Code: ErrCodeBadGateway, Message: "bad gateway"}
	ErrServiceUnavailable  = &GatewayError{Code: ErrCodeServiceUnavailable, Message: "service unavailable"}
)

func IsGatewayError(err error) bool {
	var ge *GatewayError
	return errors.As(err, &ge)
}

func GetErrorCode(err error) ErrorCode {
	var ge *GatewayError
	if errors.As(err, &ge) {
		return ge.Code
	}
	return ErrCodeInternalError
}
