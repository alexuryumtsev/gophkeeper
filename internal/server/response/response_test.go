package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	
	"github.com/uryumtsevaa/gophkeeper/internal/server/common"
)

// Вспомогательная функция для создания тестового контекста
func createTestContextForResponse() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// Тесты для NewResponseHandler
func TestNewResponseHandler(t *testing.T) {
	t.Run("создает новый ResponseHandler", func(t *testing.T) {
		handler := NewResponseHandler()

		assert.NotNil(t, handler)
		assert.Implements(t, (*common.ResponseHandler)(nil), handler)
	})
}

// Тесты для HandleError
func TestResponseHandler_HandleError(t *testing.T) {
	t.Run("обрабатывает ValidationError", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		validationErr := &common.ValidationError{
			Field:   "username",
			Message: "Username is required",
			Status:  http.StatusBadRequest,
		}

		handler.HandleError(c, validationErr)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Username is required", response["error"])
	})

	t.Run("обрабатывает ValidationError с кастомным статусом", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		validationErr := &common.ValidationError{
			Field:   "user_id",
			Message: "Unauthorized",
			Status:  http.StatusUnauthorized,
		}

		handler.HandleError(c, validationErr)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Unauthorized", response["error"])
	})

	t.Run("обрабатывает ошибку 'not found'", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		err := errors.New("user not found")

		handler.HandleError(c, err)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, jsonErr)
		assert.Equal(t, "Resource not found", response["error"])
	})

	t.Run("обрабатывает ошибку 'already exists'", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		err := errors.New("user already exists")

		handler.HandleError(c, err)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, jsonErr)
		assert.Equal(t, "Resource already exists", response["error"])
	})

	t.Run("обрабатывает ошибку 'invalid password'", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		err := errors.New("invalid password provided")

		handler.HandleError(c, err)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, jsonErr)
		assert.Equal(t, "Invalid credentials", response["error"])
	})

	t.Run("обрабатывает ошибку 'failed to decrypt'", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		err := errors.New("failed to decrypt data")

		handler.HandleError(c, err)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, jsonErr)
		assert.Equal(t, "Invalid master password", response["error"])
	})

	t.Run("обрабатывает неизвестную ошибку как internal server error", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		err := errors.New("some unknown error")

		handler.HandleError(c, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, jsonErr)
		assert.Equal(t, "Internal server error", response["error"])
	})

	t.Run("обрабатывает сложные сообщения об ошибках", func(t *testing.T) {
		testCases := []struct {
			name           string
			errorMessage   string
			expectedStatus int
			expectedError  string
		}{
			{
				name:           "ошибка с 'not found' в середине",
				errorMessage:   "record with ID 123 not found in database",
				expectedStatus: http.StatusNotFound,
				expectedError:  "Resource not found",
			},
			{
				name:           "ошибка с 'already exists' в конце",
				errorMessage:   "username testuser already exists",
				expectedStatus: http.StatusConflict,
				expectedError:  "Resource already exists",
			},
			{
				name:           "ошибка с 'invalid password' в начале",
				errorMessage:   "invalid password: must be at least 8 characters",
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "Invalid credentials",
			},
			{
				name:           "ошибка с 'failed to decrypt' вначале",
				errorMessage:   "failed to decrypt: wrong key",
				expectedStatus: http.StatusUnauthorized,
				expectedError:  "Invalid master password",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				handler := NewResponseHandler()
				c, w := createTestContextForResponse()

				err := errors.New(tc.errorMessage)
				handler.HandleError(c, err)

				assert.Equal(t, tc.expectedStatus, w.Code)

				var response map[string]interface{}
				jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, jsonErr)
				assert.Equal(t, tc.expectedError, response["error"])
			})
		}
	})
}

// Тесты для HandleSuccess
func TestResponseHandler_HandleSuccess(t *testing.T) {
	t.Run("возвращает успешный ответ с данными", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		testData := map[string]interface{}{
			"id":   123,
			"name": "test",
		}

		handler.HandleSuccess(c, http.StatusOK, testData)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(123), response["id"]) // JSON unmarshaling converts numbers to float64
		assert.Equal(t, "test", response["name"])
	})

	t.Run("возвращает успешный ответ со статусом Created", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		testData := map[string]string{
			"status": "created",
		}

		handler.HandleSuccess(c, http.StatusCreated, testData)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "created", response["status"])
	})

	t.Run("возвращает успешный ответ с nil данными", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		handler.HandleSuccess(c, http.StatusNoContent, nil)

		assert.Equal(t, http.StatusNoContent, w.Code)
		// Gin возвращает пустую строку для nil, не "null"
		assert.Contains(t, []string{"", "null\n"}, w.Body.String())
	})

	t.Run("возвращает успешный ответ с пустым объектом", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		handler.HandleSuccess(c, http.StatusOK, map[string]interface{}{})

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Empty(t, response)
	})

	t.Run("возвращает успешный ответ со сложными данными", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		testData := map[string]interface{}{
			"user": map[string]interface{}{
				"id":    "123e4567-e89b-12d3-a456-426614174000",
				"name":  "Test User",
				"email": "test@example.com",
			},
			"secrets": []map[string]interface{}{
				{
					"id":   "secret1",
					"name": "My Password",
					"type": "credentials",
				},
				{
					"id":   "secret2",
					"name": "My Note",
					"type": "text",
				},
			},
			"total": 2,
		}

		handler.HandleSuccess(c, http.StatusOK, testData)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		user := response["user"].(map[string]interface{})
		assert.Equal(t, "Test User", user["name"])
		assert.Equal(t, "test@example.com", user["email"])

		secrets := response["secrets"].([]interface{})
		assert.Len(t, secrets, 2)

		secret1 := secrets[0].(map[string]interface{})
		assert.Equal(t, "My Password", secret1["name"])
		assert.Equal(t, "credentials", secret1["type"])
	})
}

// Тесты для HandleValidationError
func TestResponseHandler_HandleValidationError(t *testing.T) {
	t.Run("возвращает ошибку валидации с полем и сообщением", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		field := "username"
		message := "Username is required"

		handler.HandleValidationError(c, field, message)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, message, response["error"])
		assert.Equal(t, field, response["field"])
	})

	t.Run("возвращает ошибку валидации для email", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		field := "email"
		message := "Invalid email format"

		handler.HandleValidationError(c, field, message)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, message, response["error"])
		assert.Equal(t, field, response["field"])
	})

	t.Run("возвращает ошибку валидации для пароля", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		field := "password"
		message := "Password must be at least 8 characters long"

		handler.HandleValidationError(c, field, message)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, message, response["error"])
		assert.Equal(t, field, response["field"])
	})

	t.Run("возвращает ошибку валидации с пустыми значениями", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		field := ""
		message := ""

		handler.HandleValidationError(c, field, message)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, message, response["error"])
		assert.Equal(t, field, response["field"])
	})

	t.Run("возвращает ошибку валидации с длинным сообщением", func(t *testing.T) {
		handler := NewResponseHandler()
		c, w := createTestContextForResponse()

		field := "description"
		message := "Description field validation failed: the provided text is too long and exceeds the maximum allowed length of 1000 characters. Please provide a shorter description."

		handler.HandleValidationError(c, field, message)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, message, response["error"])
		assert.Equal(t, field, response["field"])
	})
}