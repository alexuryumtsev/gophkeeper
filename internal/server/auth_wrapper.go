package server

// AuthServiceWrapper обертка для AuthService для реализации интерфейса роутера
type AuthServiceWrapper struct {
	*AuthService
}

// NewAuthServiceWrapper создает новую обертку
func NewAuthServiceWrapper(authService *AuthService) *AuthServiceWrapper {
	return &AuthServiceWrapper{AuthService: authService}
}

// ValidateToken проверяет JWT токен и возвращает claims как interface{}
func (w *AuthServiceWrapper) ValidateToken(tokenString string) (interface{}, error) {
	return w.AuthService.ValidateToken(tokenString)
}