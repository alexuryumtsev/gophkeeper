package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
)

// MockAuthService мок для сервиса аутентификации
type MockAuthService struct {
	mock.Mock
}

// Ensure MockAuthService implements AuthServiceInterface
var _ interfaces.AuthServiceInterface = (*MockAuthService)(nil)

func (m *MockAuthService) GenerateToken(userID uuid.UUID, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(token string) (interface{}, error) {
	args := m.Called(token)
	return args.Get(0), args.Error(1)
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authorization header required",
		},
		{
			name:           "invalid authorization header format - no bearer",
			authHeader:     "InvalidToken",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format",
		},
		{
			name:           "invalid authorization header format - wrong prefix",
			authHeader:     "Basic token123",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid authorization header format",
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid token",
		},
		{
			name:       "valid token",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "valid-token").Return(map[string]interface{}{
					"user_id": "123",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := &MockAuthService{}
			tt.mockSetup(mockAuthService)

			engine := gin.New()
			engine.Use(AuthMiddleware(mockAuthService))
			engine.GET("/test", func(c *gin.Context) {
				// Проверяем что user_id установлен в контексте для валидных токенов
				if userID, exists := c.Get("user_id"); exists {
					assert.NotNil(t, userID)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			engine.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthMiddlewareContextValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	mockAuthService.On("ValidateToken", "valid-token").Return(map[string]interface{}{
		"user_id": "test-user-123",
	}, nil)

	engine := gin.New()
	engine.Use(AuthMiddleware(mockAuthService))
	engine.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, map[string]interface{}{"user_id": "test-user-123"}, userID)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestCORSMiddlewareHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(CORSMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	// Проверяем все CORS заголовки
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, allowedHeaders, "Content-Type")
	assert.Contains(t, allowedHeaders, "Authorization")
	assert.Contains(t, allowedHeaders, "X-Master-Password")

	allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
	assert.Contains(t, allowedMethods, "POST")
	assert.Contains(t, allowedMethods, "GET")
	assert.Contains(t, allowedMethods, "PUT")
	assert.Contains(t, allowedMethods, "DELETE")
	assert.Contains(t, allowedMethods, "OPTIONS")
}

func TestCORSMiddlewareOPTIONS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(CORSMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	engine.ServeHTTP(w, req)

	// OPTIONS запрос должен вернуть 204 и не дойти до обработчика
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	// CORS заголовки должны быть установлены
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddlewareWithOtherMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("CORS with "+method, func(t *testing.T) {
			engine := gin.New()
			engine.Use(CORSMiddleware())
			engine.Handle(method, "/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/test", nil)
			engine.ServeHTTP(w, req)

			// Для не-OPTIONS запросов должны быть установлены CORS заголовки
			// и запрос должен пройти к обработчику
			assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "success")
		})
	}
}
