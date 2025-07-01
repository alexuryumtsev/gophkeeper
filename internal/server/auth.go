package server

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenExpiration = 24 * time.Hour
)

// JWTClaims кастомные claims для JWT токена
type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// AuthServiceInterface интерфейс для работы с аутентификацией
type AuthServiceInterface interface {
	GenerateToken(userID uuid.UUID, username string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}

// AuthService сервис для работы с аутентификацией
type AuthService struct {
	secretKey []byte
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(secretKey string) *AuthService {
	return &AuthService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken генерирует JWT токен для пользователя
func (s *AuthService) GenerateToken(userID uuid.UUID, username string) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gophkeeper",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken проверяет JWT токен и возвращает claims
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
