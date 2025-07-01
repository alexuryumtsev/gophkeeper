package interfaces

// AuthServiceInterface интерфейс для сервиса авторизации
type AuthServiceInterface interface {
	ValidateToken(token string) (interface{}, error)
}