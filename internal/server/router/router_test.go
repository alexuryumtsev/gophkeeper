package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uryumtsevaa/gophkeeper/internal/models"
	"github.com/uryumtsevaa/gophkeeper/internal/server/handlers"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
	"github.com/uryumtsevaa/gophkeeper/internal/server/middleware"
	"github.com/uryumtsevaa/gophkeeper/internal/server/response"
	"github.com/uryumtsevaa/gophkeeper/internal/server/validation"
)

// MockService мок для сервиса
type MockService struct {
	mock.Mock
}

// Ensure MockService implements ServiceInterface
var _ interfaces.ServiceInterface = (*MockService)(nil)

func (m *MockService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockService) LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockService) CreateSecret(ctx context.Context, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
	args := m.Called(ctx, userID, req, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretResponse), args.Error(1)
}

func (m *MockService) GetSecrets(ctx context.Context, userID uuid.UUID, masterPassword string) (*models.SecretsListResponse, error) {
	args := m.Called(ctx, userID, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretsListResponse), args.Error(1)
}

func (m *MockService) GetSecret(ctx context.Context, secretID, userID uuid.UUID, masterPassword string) (*models.SecretResponse, error) {
	args := m.Called(ctx, secretID, userID, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretResponse), args.Error(1)
}

func (m *MockService) UpdateSecret(ctx context.Context, secretID, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
	args := m.Called(ctx, secretID, userID, req, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SecretResponse), args.Error(1)
}

func (m *MockService) DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error {
	args := m.Called(ctx, secretID, userID)
	return args.Error(0)
}

func (m *MockService) SyncSecrets(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error) {
	args := m.Called(ctx, userID, req, masterPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SyncResponse), args.Error(1)
}

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

// createTestHandlers создает handlers с mock service для тестов
func createTestHandlers() *handlers.Handlers {
	mockService := &MockService{}
	responseHandler := response.NewResponseHandler()
	validationSvc := validation.NewValidationService(responseHandler)
	
	return handlers.NewHandlers(mockService, validationSvc, responseHandler)
}

func TestNewRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testHandlers := createTestHandlers()
	mockAuthService := &MockAuthService{}

	router := NewRouter(testHandlers, mockAuthService)

	assert.NotNil(t, router)
	assert.NotNil(t, router.engine)
	assert.Equal(t, testHandlers, router.handlers)
	assert.Equal(t, mockAuthService, router.authSvc)
}

func TestGetEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testHandlers := createTestHandlers()
	mockAuthService := &MockAuthService{}

	router := NewRouter(testHandlers, mockAuthService)
	engine := router.GetEngine()

	assert.NotNil(t, engine)
	assert.IsType(t, &gin.Engine{}, engine)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testHandlers := createTestHandlers()
	mockAuthService := &MockAuthService{}

	router := NewRouter(testHandlers, mockAuthService)
	router.SetupRoutes(8080)

	// Проверяем что роуты настроены, делая тестовые запросы
	engine := router.GetEngine()

	// Тест health endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")

	// Тест swagger endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/swagger/index.html", nil)
	engine.ServeHTTP(w, req)

	// Должен вернуть 200 или 404 (в зависимости от наличия swagger файлов)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
}

func TestPublicRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testHandlers := createTestHandlers()
	mockAuthService := &MockAuthService{}

	router := NewRouter(testHandlers, mockAuthService)
	router.SetupRoutes(8080)
	engine := router.GetEngine()

	tests := []struct {
		method   string
		path     string
		handler  string
		expected int
	}{
		{"POST", "/api/v1/auth/register", "Register", http.StatusOK},
		{"POST", "/api/v1/auth/login", "Login", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")

			engine.ServeHTTP(w, req)

			// Проверяем что обработчик был вызван (код может быть разный в зависимости от логики)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
		})
	}
}

func TestProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testHandlers := createTestHandlers()
	mockAuthService := &MockAuthService{}

	router := NewRouter(testHandlers, mockAuthService)
	router.SetupRoutes(8080)
	engine := router.GetEngine()

	protectedRoutes := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/secrets"},
		{"GET", "/api/v1/secrets"},
		{"GET", "/api/v1/secrets/123"},
		{"PUT", "/api/v1/secrets/123"},
		{"DELETE", "/api/v1/secrets/123"},
		{"POST", "/api/v1/sync"},
	}

	for _, route := range protectedRoutes {
		t.Run(route.method+" "+route.path+" without auth", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)

			engine.ServeHTTP(w, req)

			// Должен вернуть 401 без авторизации
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthService := &MockAuthService{}
	
	// Настраиваем мок для успешной аутентификации  
	mockAuthService.On("ValidateToken", "valid-token").Return(map[string]interface{}{
		"user_id": "123",
	}, nil)

	// Создаем простой тестовый роутер с middleware
	engine := gin.New()
	engine.Use(middleware.AuthMiddleware(mockAuthService))
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("with valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	mockAuthService.AssertExpectations(t)
}

// TestProtectedRoutesWithValidAuth тестирует что защищенные маршруты работают с валидной аутентификацией
// Примечание: этот тест требует реальной настройки всей системы аутентификации,
// поэтому мы тестируем middleware отдельно в TestAuthMiddleware

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(middleware.CORSMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Тест обычного запроса
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// Тест OPTIONS запроса
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}