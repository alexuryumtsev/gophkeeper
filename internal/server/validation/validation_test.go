package validation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/uryumtsevaa/gophkeeper/internal/server/common"
)

// MockResponseHandler для тестирования ValidationService
type MockResponseHandlerForValidation struct {
	mock.Mock
}

func (m *MockResponseHandlerForValidation) HandleError(c *gin.Context, err error) {
	m.Called(c, err)
}

func (m *MockResponseHandlerForValidation) HandleSuccess(c *gin.Context, statusCode int, data interface{}) {
	m.Called(c, statusCode, data)
}

func (m *MockResponseHandlerForValidation) HandleValidationError(c *gin.Context, field, message string) {
	m.Called(c, field, message)
}

// Вспомогательная функция для создания gin context с записанным ответом
func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// Тесты для NewValidationService
func TestNewValidationService(t *testing.T) {
	t.Run("создает новый ValidationService", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		assert.NotNil(t, service)
		assert.Implements(t, (*common.ValidationService)(nil), service)
	})
}

// Тесты для ValidateUserID
func TestValidationService_ValidateUserID(t *testing.T) {
	t.Run("успешная валидация валидного UUID", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		testUUID := uuid.New()
		c.Set("user_id", testUUID.String())

		result, err := service.ValidateUserID(c)

		assert.NoError(t, err)
		assert.Equal(t, testUUID, result)
		mockResponseHandler.AssertNotCalled(t, "HandleError")
	})

	t.Run("ошибка при отсутствии user_id", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		// Не устанавливаем user_id

		expectedErr := &common.ValidationError{
			Field:   "user_id",
			Message: "Unauthorized",
			Status:  http.StatusUnauthorized,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateUserID(c)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
		assert.Equal(t, "Unauthorized", err.Error())
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном UUID", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Set("user_id", "invalid-uuid")

		expectedErr := &common.ValidationError{
			Field:   "user_id",
			Message: "Invalid user ID",
			Status:  http.StatusBadRequest,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateUserID(c)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
		assert.Equal(t, "Invalid user ID", err.Error())
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при пустом user_id", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Set("user_id", "")

		expectedErr := &common.ValidationError{
			Field:   "user_id",
			Message: "Unauthorized",
			Status:  http.StatusUnauthorized,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateUserID(c)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
		mockResponseHandler.AssertExpectations(t)
	})
}

// Тесты для ValidateMasterPassword
func TestValidationService_ValidateMasterPassword(t *testing.T) {
	t.Run("успешная валидация мастер-пароля", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", nil)
		c.Request.Header.Set("X-Master-Password", "test-master-password")

		result, err := service.ValidateMasterPassword(c)

		assert.NoError(t, err)
		assert.Equal(t, "test-master-password", result)
		mockResponseHandler.AssertNotCalled(t, "HandleError")
	})

	t.Run("ошибка при отсутствии мастер-пароля", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", nil)
		// Не устанавливаем заголовок X-Master-Password

		expectedErr := &common.ValidationError{
			Field:   "master_password",
			Message: "Master password required",
			Status:  http.StatusBadRequest,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateMasterPassword(c)

		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Equal(t, "Master password required", err.Error())
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при пустом мастер-пароле", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", nil)
		c.Request.Header.Set("X-Master-Password", "")

		expectedErr := &common.ValidationError{
			Field:   "master_password",
			Message: "Master password required",
			Status:  http.StatusBadRequest,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateMasterPassword(c)

		assert.Error(t, err)
		assert.Equal(t, "", result)
		mockResponseHandler.AssertExpectations(t)
	})
}

// Тесты для ValidateSecretID
func TestValidationService_ValidateSecretID(t *testing.T) {
	t.Run("успешная валидация валидного secret ID", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		testUUID := uuid.New()
		c.Params = gin.Params{
			{Key: "id", Value: testUUID.String()},
		}

		result, err := service.ValidateSecretID(c)

		assert.NoError(t, err)
		assert.Equal(t, testUUID, result)
		mockResponseHandler.AssertNotCalled(t, "HandleError")
	})

	t.Run("ошибка при отсутствии secret ID", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		// Не устанавливаем параметр id

		expectedErr := &common.ValidationError{
			Field:   "secret_id",
			Message: "Secret ID is required",
			Status:  http.StatusBadRequest,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateSecretID(c)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
		assert.Equal(t, "Secret ID is required", err.Error())
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном secret ID", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Params = gin.Params{
			{Key: "id", Value: "invalid-uuid"},
		}

		expectedErr := &common.ValidationError{
			Field:   "secret_id",
			Message: "Invalid secret ID",
			Status:  http.StatusBadRequest,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateSecretID(c)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
		assert.Equal(t, "Invalid secret ID", err.Error())
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при пустом secret ID", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		c, _ := createTestContext()
		c.Params = gin.Params{
			{Key: "id", Value: ""},
		}

		expectedErr := &common.ValidationError{
			Field:   "secret_id",
			Message: "Secret ID is required",
			Status:  http.StatusBadRequest,
		}

		mockResponseHandler.On("HandleError", c, expectedErr).Return()

		result, err := service.ValidateSecretID(c)

		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, result)
		mockResponseHandler.AssertExpectations(t)
	})
}

// Тесты для BindJSON
func TestValidationService_BindJSON(t *testing.T) {
	t.Run("успешное связывание JSON", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		jsonData := `{"name":"test","age":25}`
		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(jsonData))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestStruct
		err := service.BindJSON(c, &result)

		assert.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 25, result.Age)
		mockResponseHandler.AssertNotCalled(t, "HandleError")
	})

	t.Run("ошибка при невалидном JSON", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		invalidJSON := `{"name":"test","age":}`
		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(invalidJSON))
		c.Request.Header.Set("Content-Type", "application/json")

		mockResponseHandler.On("HandleError", c, mock.MatchedBy(func(err error) bool {
			validationErr, ok := err.(*common.ValidationError)
			return ok && validationErr.Field == "json_body"
		})).Return()

		var result TestStruct
		err := service.BindJSON(c, &result)

		assert.Error(t, err)
		validationErr, ok := err.(*common.ValidationError)
		assert.True(t, ok)
		assert.Equal(t, "json_body", validationErr.Field)
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при пустом теле запроса", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		type TestStruct struct {
			Name string `json:"name" binding:"required"`
			Age  int    `json:"age" binding:"required"`
		}

		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("{}"))
		c.Request.Header.Set("Content-Type", "application/json")

		mockResponseHandler.On("HandleError", c, mock.MatchedBy(func(err error) bool {
			validationErr, ok := err.(*common.ValidationError)
			return ok && validationErr.Field == "json_body"
		})).Return()

		var result TestStruct
		err := service.BindJSON(c, &result)

		assert.Error(t, err)
		validationErr, ok := err.(*common.ValidationError)
		assert.True(t, ok)
		assert.Equal(t, "json_body", validationErr.Field)
		mockResponseHandler.AssertExpectations(t)
	})

	t.Run("ошибка при неподдерживаемом Content-Type", func(t *testing.T) {
		mockResponseHandler := new(MockResponseHandlerForValidation)
		service := NewValidationService(mockResponseHandler)

		type TestStruct struct {
			Name string `json:"name"`
		}

		c, _ := createTestContext()
		c.Request, _ = http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("name=test"))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		mockResponseHandler.On("HandleError", c, mock.MatchedBy(func(err error) bool {
			validationErr, ok := err.(*common.ValidationError)
			return ok && validationErr.Field == "json_body"
		})).Return()

		var result TestStruct
		err := service.BindJSON(c, &result)

		assert.Error(t, err)
		validationErr, ok := err.(*common.ValidationError)
		assert.True(t, ok)
		assert.Equal(t, "json_body", validationErr.Field)
		mockResponseHandler.AssertExpectations(t)
	})
}