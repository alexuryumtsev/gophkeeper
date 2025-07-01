package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHandlers мок для обработчиков
type MockHandlers struct {
	mock.Mock
}

func (m *MockHandlers) Register(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) Login(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) CreateSecret(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) GetSecrets(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) GetSecret(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) UpdateSecret(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) DeleteSecret(c *gin.Context) {
	m.Called(c)
}

func (m *MockHandlers) SyncSecrets(c *gin.Context) {
	m.Called(c)
}

// MockAuthService мок для сервиса аутентификации
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(token string) (interface{}, error) {
	args := m.Called(token)
	return args.Get(0), args.Error(1)
}

func TestNewRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandlers := &MockHandlers{}
	mockAuthService := &MockAuthService{}

	router := NewRouter(mockHandlers, mockAuthService)

	assert.NotNil(t, router)
	assert.NotNil(t, router.engine)
	assert.Equal(t, mockHandlers, router.handlers)
	assert.Equal(t, mockAuthService, router.authSvc)
}

func TestGetEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandlers := &MockHandlers{}
	mockAuthService := &MockAuthService{}

	router := NewRouter(mockHandlers, mockAuthService)
	engine := router.GetEngine()

	assert.NotNil(t, engine)
	assert.IsType(t, &gin.Engine{}, engine)
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandlers := &MockHandlers{}
	mockAuthService := &MockAuthService{}

	router := NewRouter(mockHandlers, mockAuthService)
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

	mockHandlers := &MockHandlers{}
	mockAuthService := &MockAuthService{}

	// Настраиваем мок для ожидания вызовов
	mockHandlers.On("Register", mock.AnythingOfType("*gin.Context")).Return()
	mockHandlers.On("Login", mock.AnythingOfType("*gin.Context")).Return()

	router := NewRouter(mockHandlers, mockAuthService)
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

	mockHandlers.AssertExpectations(t)
}

func TestProtectedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandlers := &MockHandlers{}
	mockAuthService := &MockAuthService{}

	router := NewRouter(mockHandlers, mockAuthService)
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

func TestProtectedRoutesWithValidAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockHandlers := &MockHandlers{}
	mockAuthService := &MockAuthService{}

	// Настраиваем мок для успешной аутентификации
	mockAuthService.On("ValidateToken", "valid-token").Return(map[string]interface{}{
		"user_id": "123",
	}, nil)

	// Настраиваем моки для обработчиков
	mockHandlers.On("CreateSecret", mock.AnythingOfType("*gin.Context")).Return()
	mockHandlers.On("GetSecrets", mock.AnythingOfType("*gin.Context")).Return()
	mockHandlers.On("GetSecret", mock.AnythingOfType("*gin.Context")).Return()
	mockHandlers.On("UpdateSecret", mock.AnythingOfType("*gin.Context")).Return()
	mockHandlers.On("DeleteSecret", mock.AnythingOfType("*gin.Context")).Return()
	mockHandlers.On("SyncSecrets", mock.AnythingOfType("*gin.Context")).Return()

	router := NewRouter(mockHandlers, mockAuthService)
	router.SetupRoutes(8080)
	engine := router.GetEngine()

	protectedRoutes := []struct {
		method  string
		path    string
		handler string
	}{
		{"POST", "/api/v1/secrets", "CreateSecret"},
		{"GET", "/api/v1/secrets", "GetSecrets"},
		{"GET", "/api/v1/secrets/123", "GetSecret"},
		{"PUT", "/api/v1/secrets/123", "UpdateSecret"},
		{"DELETE", "/api/v1/secrets/123", "DeleteSecret"},
		{"POST", "/api/v1/sync", "SyncSecrets"},
	}

	for _, route := range protectedRoutes {
		t.Run(route.method+" "+route.path+" with auth", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(route.method, route.path, nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("Content-Type", "application/json")

			engine.ServeHTTP(w, req)

			// Не должен вернуть 401 с валидной авторизацией
			assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		})
	}

	mockAuthService.AssertExpectations(t)
	mockHandlers.AssertExpectations(t)
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(CORSMiddleware())
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