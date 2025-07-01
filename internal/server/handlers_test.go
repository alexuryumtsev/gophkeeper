package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// MockService для тестирования handlers
type MockService struct {
	mock.Mock
}

// Ensure MockService implements ServiceInterface
var _ ServiceInterface = (*MockService)(nil)

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

// MockValidationService для тестирования
type MockValidationService struct {
	mock.Mock
}

func (m *MockValidationService) ValidateUserID(c *gin.Context) (uuid.UUID, error) {
	args := m.Called(c)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockValidationService) ValidateMasterPassword(c *gin.Context) (string, error) {
	args := m.Called(c)
	return args.String(0), args.Error(1)
}

func (m *MockValidationService) ValidateSecretID(c *gin.Context) (uuid.UUID, error) {
	args := m.Called(c)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockValidationService) BindJSON(c *gin.Context, v interface{}) error {
	args := m.Called(c, v)
	if args.Error(0) == nil {
		// Simulate successful binding by copying test data
		if registerReq, ok := v.(*models.RegisterRequest); ok && args.Get(1) != nil {
			*registerReq = *args.Get(1).(*models.RegisterRequest)
		}
		if loginReq, ok := v.(*models.LoginRequest); ok && args.Get(1) != nil {
			*loginReq = *args.Get(1).(*models.LoginRequest)
		}
		if secretReq, ok := v.(*models.SecretRequest); ok && args.Get(1) != nil {
			*secretReq = *args.Get(1).(*models.SecretRequest)
		}
		if syncReq, ok := v.(*models.SyncRequest); ok && args.Get(1) != nil {
			*syncReq = *args.Get(1).(*models.SyncRequest)
		}
	}
	return args.Error(0)
}

// MockResponseHandler для тестирования
type MockResponseHandler struct {
	mock.Mock
}

func (m *MockResponseHandler) HandleError(c *gin.Context, err error) {
	m.Called(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func (m *MockResponseHandler) HandleSuccess(c *gin.Context, statusCode int, data interface{}) {
	m.Called(c, statusCode, data)
	c.JSON(statusCode, data)
}

func (m *MockResponseHandler) HandleValidationError(c *gin.Context, field, message string) {
	m.Called(c, field, message)
	c.JSON(http.StatusBadRequest, gin.H{"error": message, "field": field})
}

// Тестовые данные
func getTestUser() models.User {
	return models.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func getTestRegisterRequest() *models.RegisterRequest {
	return &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
}

func getTestLoginRequest() *models.LoginRequest {
	return &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
}

func getTestSecretRequest() *models.SecretRequest {
	return &models.SecretRequest{
		Type: models.SecretTypeCredentials,
		Name: "Test Credentials",
		Data: models.Credentials{
			Username: "user",
			Password: "pass",
			URL:      "https://example.com",
		},
		Metadata: "test metadata",
	}
}

// Тестирование Register handler
func TestHandlers_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUser := getTestUser()
	testRegisterReq := getTestRegisterRequest()

	t.Run("successful registration", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.LoginResponse{
			Token: "test-token",
			User:  testUser,
		}

		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(nil, testRegisterReq)
		mockService.On("RegisterUser", mock.Anything, testRegisterReq).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusCreated, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testRegisterReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.Register(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		bindError := fmt.Errorf("invalid JSON")
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(bindError, nil)

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.Register(c)

		mockValidation.AssertExpectations(t)
		mockService.AssertNotCalled(t, "RegisterUser")
	})

	t.Run("service error", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		serviceError := fmt.Errorf("user already exists")
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(nil, testRegisterReq)
		mockService.On("RegisterUser", mock.Anything, testRegisterReq).Return(nil, serviceError)
		mockResponse.On("HandleError", mock.Anything, serviceError).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testRegisterReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.Register(c)

		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})
}

// Тестирование Login handler
func TestHandlers_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUser := getTestUser()
	testLoginReq := getTestLoginRequest()

	t.Run("successful login", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.LoginResponse{
			Token: "test-token",
			User:  testUser,
		}

		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, testLoginReq)
		mockService.On("LoginUser", mock.Anything, testLoginReq).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusOK, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testLoginReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		authError := fmt.Errorf("invalid password")
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, testLoginReq)
		mockService.On("LoginUser", mock.Anything, testLoginReq).Return(nil, authError)
		mockResponse.On("HandleError", mock.Anything, authError).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testLoginReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.Login(c)

		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})
}

// Тестирование CreateSecret handler
func TestHandlers_CreateSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testSecretReq := getTestSecretRequest()
	testUserID := uuid.New()
	testMasterPassword := "master-password"

	t.Run("successful secret creation", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.SecretResponse{
			ID:        uuid.New(),
			Type:      testSecretReq.Type,
			Name:      testSecretReq.Name,
			Data:      testSecretReq.Data,
			Metadata:  testSecretReq.Metadata,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			SyncHash:  "test-hash",
		}

		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.SecretRequest")).Return(nil, testSecretReq)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("CreateSecret", mock.Anything, testUserID, testSecretReq, testMasterPassword).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusCreated, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testSecretReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/secrets", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.CreateSecret(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("validation error - invalid user ID", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		validationError := fmt.Errorf("unauthorized")
		mockValidation.On("ValidateUserID", mock.Anything).Return(uuid.Nil, validationError)

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testSecretReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/secrets", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.CreateSecret(c)

		mockValidation.AssertExpectations(t)
		mockService.AssertNotCalled(t, "CreateSecret")
	})

	t.Run("validation error - missing master password", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		masterPasswordError := fmt.Errorf("master password required")
		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.SecretRequest")).Return(nil, testSecretReq)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return("", masterPasswordError)

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testSecretReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/secrets", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.CreateSecret(c)

		mockValidation.AssertExpectations(t)
		mockService.AssertNotCalled(t, "CreateSecret")
	})
}

// Тестирование GetSecrets handler
func TestHandlers_GetSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	testMasterPassword := "master-password"

	t.Run("successful get secrets", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.SecretsListResponse{
			Secrets: []models.SecretResponse{
				{
					ID:   uuid.New(),
					Type: models.SecretTypeCredentials,
					Name: "Test Secret 1",
				},
				{
					ID:   uuid.New(),
					Type: models.SecretTypeText,
					Name: "Test Secret 2",
				},
			},
			Total: 2,
		}

		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("GetSecrets", mock.Anything, testUserID, testMasterPassword).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusOK, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/secrets", nil)

		handlers.GetSecrets(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		validationError := fmt.Errorf("unauthorized")
		mockValidation.On("ValidateUserID", mock.Anything).Return(uuid.Nil, validationError)

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/secrets", nil)

		handlers.GetSecrets(c)

		mockValidation.AssertExpectations(t)
		mockService.AssertNotCalled(t, "GetSecrets")
	})
}

// Тестирование GetSecret handler
func TestHandlers_GetSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	testSecretID := uuid.New()
	testMasterPassword := "master-password"

	t.Run("successful get secret", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.SecretResponse{
			ID:       testSecretID,
			Type:     models.SecretTypeCredentials,
			Name:     "Test Secret",
			Data:     models.Credentials{Username: "user", Password: "pass"},
			Metadata: "test",
		}

		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("ValidateSecretID", mock.Anything).Return(testSecretID, nil)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("GetSecret", mock.Anything, testSecretID, testUserID, testMasterPassword).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusOK, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/secrets/%s", testSecretID.String()), nil)

		handlers.GetSecret(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("secret not found", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		notFoundError := fmt.Errorf("secret not found")
		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("ValidateSecretID", mock.Anything).Return(testSecretID, nil)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("GetSecret", mock.Anything, testSecretID, testUserID, testMasterPassword).Return(nil, notFoundError)
		mockResponse.On("HandleError", mock.Anything, notFoundError).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/secrets/%s", testSecretID.String()), nil)

		handlers.GetSecret(c)

		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})
}

// Тестирование UpdateSecret handler
func TestHandlers_UpdateSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	testSecretID := uuid.New()
	testMasterPassword := "master-password"
	testSecretReq := getTestSecretRequest()

	t.Run("successful update secret", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.SecretResponse{
			ID:       testSecretID,
			Type:     testSecretReq.Type,
			Name:     testSecretReq.Name,
			Data:     testSecretReq.Data,
			Metadata: testSecretReq.Metadata,
		}

		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("ValidateSecretID", mock.Anything).Return(testSecretID, nil)
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.SecretRequest")).Return(nil, testSecretReq)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("UpdateSecret", mock.Anything, testSecretID, testUserID, testSecretReq, testMasterPassword).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusOK, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testSecretReq)
		c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/secrets/%s", testSecretID.String()), bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.UpdateSecret(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})
}

// Тестирование DeleteSecret handler
func TestHandlers_DeleteSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	testSecretID := uuid.New()

	t.Run("successful delete secret", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("ValidateSecretID", mock.Anything).Return(testSecretID, nil)
		mockService.On("DeleteSecret", mock.Anything, testSecretID, testUserID).Return(nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusNoContent, nil).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/secrets/%s", testSecretID.String()), nil)

		handlers.DeleteSecret(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("secret not found", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		notFoundError := fmt.Errorf("secret not found")
		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("ValidateSecretID", mock.Anything).Return(testSecretID, nil)
		mockService.On("DeleteSecret", mock.Anything, testSecretID, testUserID).Return(notFoundError)
		mockResponse.On("HandleError", mock.Anything, notFoundError).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/secrets/%s", testSecretID.String()), nil)

		handlers.DeleteSecret(c)

		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})
}

// Тестирование SyncSecrets handler
func TestHandlers_SyncSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	testMasterPassword := "master-password"
	testSyncReq := &models.SyncRequest{
		LastSyncTime: time.Now().Add(-time.Hour),
		ClientHashes: map[uuid.UUID]string{
			uuid.New(): "hash1",
			uuid.New(): "hash2",
		},
	}

	t.Run("successful sync", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		expectedResponse := &models.SyncResponse{
			UpdatedSecrets: []models.SecretResponse{
				{ID: uuid.New(), Name: "Updated Secret"},
			},
			DeletedSecrets: []uuid.UUID{uuid.New()},
			SyncTime:       time.Now(),
		}

		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.SyncRequest")).Return(nil, testSyncReq)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("SyncSecrets", mock.Anything, testUserID, testSyncReq, testMasterPassword).Return(expectedResponse, nil)
		mockResponse.On("HandleSuccess", mock.Anything, http.StatusOK, expectedResponse).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testSyncReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/sync", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.SyncSecrets(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})

	t.Run("sync service error", func(t *testing.T) {
		mockService := new(MockService)
		mockValidation := new(MockValidationService)
		mockResponse := new(MockResponseHandler)

		syncError := fmt.Errorf("sync failed")
		mockValidation.On("ValidateUserID", mock.Anything).Return(testUserID, nil)
		mockValidation.On("BindJSON", mock.Anything, mock.AnythingOfType("*models.SyncRequest")).Return(nil, testSyncReq)
		mockValidation.On("ValidateMasterPassword", mock.Anything).Return(testMasterPassword, nil)
		mockService.On("SyncSecrets", mock.Anything, testUserID, testSyncReq, testMasterPassword).Return(nil, syncError)
		mockResponse.On("HandleError", mock.Anything, syncError).Return()

		handlers := NewHandlers(mockService, mockValidation, mockResponse)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody, _ := json.Marshal(testSyncReq)
		c.Request, _ = http.NewRequest(http.MethodPost, "/sync", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handlers.SyncSecrets(c)

		mockService.AssertExpectations(t)
		mockValidation.AssertExpectations(t)
		mockResponse.AssertExpectations(t)
	})
}