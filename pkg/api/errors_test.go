package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:    400,
		Message: "Bad Request",
		Details: "Invalid input data",
	}

	expected := "Bad Request"
	assert.Equal(t, expected, err.Error())
}

func TestAPIError_IsClientError(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{399, false},
		{400, true},
		{404, true},
		{499, true},
		{500, false},
		{600, false},
	}

	for _, tt := range tests {
		err := &APIError{Code: tt.code}
		assert.Equal(t, tt.expected, err.IsClientError())
	}
}

func TestAPIError_IsServerError(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{499, false},
		{500, true},
		{502, true},
		{599, true},
		{600, true},
	}

	for _, tt := range tests {
		err := &APIError{Code: tt.code}
		assert.Equal(t, tt.expected, err.IsServerError())
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "username",
		Message: "is required",
	}

	expected := "validation error for field 'username': is required"
	assert.Equal(t, expected, err.Error())
}

func TestAuthError_Error(t *testing.T) {
	err := &AuthError{
		Message: "Invalid credentials",
	}

	expected := "Invalid credentials"
	assert.Equal(t, expected, err.Error())
}

func TestNetworkError_Error(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := &NetworkError{
		Operation: "GET",
		URL:       "/api/v1/secrets",
		Err:       originalErr,
	}

	expected := "network error during GET to /api/v1/secrets: connection timeout"
	assert.Equal(t, expected, err.Error())
}

func TestNetworkError_Unwrap(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := &NetworkError{
		Operation: "GET",
		URL:       "/api/v1/secrets",
		Err:       originalErr,
	}

	assert.Equal(t, originalErr, err.Unwrap())
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(404, "Not Found")

	assert.Equal(t, 404, err.Code)
	assert.Equal(t, "Not Found", err.Message)
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "invalid format")

	assert.Equal(t, "email", err.Field)
	assert.Equal(t, "invalid format", err.Message)
}

func TestNewAuthError(t *testing.T) {
	err := NewAuthError("Token expired")

	assert.Equal(t, "Token expired", err.Message)
}

func TestNewNetworkError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := NewNetworkError("POST", "/api/v1/auth/login", originalErr)

	assert.Equal(t, "POST", err.Operation)
	assert.Equal(t, "/api/v1/auth/login", err.URL)
	assert.Equal(t, originalErr, err.Err)
}

func TestErrorChaining(t *testing.T) {
	originalErr := errors.New("connection failed")
	networkErr := NewNetworkError("GET", "/api", originalErr)

	// Проверяем что errors.Is работает
	assert.True(t, errors.Is(networkErr, originalErr))

	// Проверяем что errors.As работает
	var netErr *NetworkError
	assert.True(t, errors.As(networkErr, &netErr))
	assert.Equal(t, "GET", netErr.Operation)
	assert.Equal(t, "/api", netErr.URL)
}
