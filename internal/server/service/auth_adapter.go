package service

import (
	"github.com/google/uuid"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
	"github.com/uryumtsevaa/gophkeeper/internal/server/auth"
)

// authServiceAdapter адаптер для auth.AuthService
type authServiceAdapter struct {
	authService *auth.AuthService
}

// NewAuthServiceAdapter создает новый адаптер для auth service
func NewAuthServiceAdapter(authService *auth.AuthService) interfaces.AuthServiceInterface {
	return &authServiceAdapter{
		authService: authService,
	}
}

// GenerateToken генерирует токен
func (a *authServiceAdapter) GenerateToken(userID uuid.UUID, username string) (string, error) {
	return a.authService.GenerateToken(userID, username)
}

// ValidateToken валидирует токен и возвращает результат как interface{}
func (a *authServiceAdapter) ValidateToken(token string) (interface{}, error) {
	return a.authService.ValidateToken(token)
}