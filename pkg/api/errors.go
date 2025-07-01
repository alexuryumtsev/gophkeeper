package api

import (
	"fmt"
	"net/http"
)

// APIError представляет ошибку API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Field   string `json:"field,omitempty"`
}

// Error реализует интерфейс error
func (e *APIError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s (field: %s)", e.Message, e.Field)
	}
	return e.Message
}

// IsClientError проверяет, является ли ошибка клиентской (4xx)
func (e *APIError) IsClientError() bool {
	return e.Code >= 400 && e.Code < 500
}

// IsServerError проверяет, является ли ошибка серверной (5xx)
func (e *APIError) IsServerError() bool {
	return e.Code >= 500
}

// ValidationError ошибка валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error реализует интерфейс error
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// AuthError ошибка аутентификации
type AuthError struct {
	Message string `json:"message"`
}

// Error реализует интерфейс error
func (e *AuthError) Error() string {
	return e.Message
}

// NetworkError ошибка сети
type NetworkError struct {
	Operation string `json:"operation"`
	URL       string `json:"url"`
	Err       error  `json:"-"`
}

// Error реализует интерфейс error
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error during %s to %s: %v", e.Operation, e.URL, e.Err)
}

// Unwrap возвращает обернутую ошибку
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// Предопределенные ошибки

// NewAPIError создает новую API ошибку
func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewValidationError создает новую ошибку валидации
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewAuthError создает новую ошибку аутентификации
func NewAuthError(message string) *AuthError {
	return &AuthError{
		Message: message,
	}
}

// NewNetworkError создает новую сетевую ошибку
func NewNetworkError(operation, url string, err error) *NetworkError {
	return &NetworkError{
		Operation: operation,
		URL:       url,
		Err:       err,
	}
}

// Общие ошибки

var (
	// ErrUnauthorized ошибка неавторизованного доступа
	ErrUnauthorized = NewAPIError(http.StatusUnauthorized, "Unauthorized")
	
	// ErrForbidden ошибка запрещенного доступа
	ErrForbidden = NewAPIError(http.StatusForbidden, "Forbidden")
	
	// ErrNotFound ошибка не найденного ресурса
	ErrNotFound = NewAPIError(http.StatusNotFound, "Resource not found")
	
	// ErrConflict ошибка конфликта (ресурс уже существует)
	ErrConflict = NewAPIError(http.StatusConflict, "Resource already exists")
	
	// ErrBadRequest ошибка неверного запроса
	ErrBadRequest = NewAPIError(http.StatusBadRequest, "Bad request")
	
	// ErrInternalServer внутренняя ошибка сервера
	ErrInternalServer = NewAPIError(http.StatusInternalServerError, "Internal server error")
	
	// ErrInvalidCredentials ошибка неверных учетных данных
	ErrInvalidCredentials = NewAuthError("Invalid username or password")
	
	// ErrInvalidToken ошибка неверного токена
	ErrInvalidToken = NewAuthError("Invalid or expired token")
	
	// ErrMasterPasswordRequired ошибка отсутствия мастер-пароля
	ErrMasterPasswordRequired = NewAPIError(http.StatusBadRequest, "Master password is required")
	
	// ErrInvalidMasterPassword ошибка неверного мастер-пароля
	ErrInvalidMasterPassword = NewAPIError(http.StatusUnauthorized, "Invalid master password")
)

// IsAPIError проверяет, является ли ошибка API ошибкой
func IsAPIError(err error) (*APIError, bool) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, true
	}
	return nil, false
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации
func IsValidationError(err error) (*ValidationError, bool) {
	if valErr, ok := err.(*ValidationError); ok {
		return valErr, true
	}
	return nil, false
}

// IsAuthError проверяет, является ли ошибка ошибкой аутентификации
func IsAuthError(err error) (*AuthError, bool) {
	if authErr, ok := err.(*AuthError); ok {
		return authErr, true
	}
	return nil, false
}

// IsNetworkError проверяет, является ли ошибка сетевой ошибкой
func IsNetworkError(err error) (*NetworkError, bool) {
	if netErr, ok := err.(*NetworkError); ok {
		return netErr, true
	}
	return nil, false
}