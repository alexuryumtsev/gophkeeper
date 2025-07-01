package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

func TestNewClient(t *testing.T) {
	baseURL := "https://api.example.com"
	client := NewClient(baseURL)

	assert.NotNil(t, client)
	assert.Equal(t, baseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Empty(t, client.token)
	assert.Empty(t, client.masterPassword)
}

func TestClient_SetToken(t *testing.T) {
	client := NewClient("https://api.example.com")
	token := "test-token-123"

	client.SetToken(token)
	assert.Equal(t, token, client.token)
}

func TestClient_SetMasterPassword(t *testing.T) {
	client := NewClient("https://api.example.com")
	password := "master-password-123"

	client.SetMasterPassword(password)
	assert.Equal(t, password, client.masterPassword)
}

func TestClient_Register(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/auth/register", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Проверяем тело запроса
		var req models.RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "password123", req.Password)

		// Возвращаем успешный ответ
		userID := uuid.New()
		response := models.LoginResponse{
			Token: "test-token",
			User: models.User{
				ID:       userID,
				Username: "testuser",
				Email:    "test@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	response, err := client.Register(ctx, "testuser", "test@example.com", "password123")
	require.NoError(t, err)
	assert.Equal(t, "test-token", response.Token)
	assert.Equal(t, "testuser", response.User.Username)
}

func TestClient_Login(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/auth/login", r.URL.Path)

		var req models.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "password123", req.Password)

		userID := uuid.New()
		response := models.LoginResponse{
			Token: "login-token",
			User: models.User{
				ID:       userID,
				Username: "testuser",
				Email:    "test@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	response, err := client.Login(ctx, "testuser", "password123")
	require.NoError(t, err)
	assert.Equal(t, "login-token", response.Token)
	assert.Equal(t, "testuser", response.User.Username)
}

func TestClient_CreateSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/secrets", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "master-password", r.Header.Get("X-Master-Password"))

		var req models.SecretRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "Test Secret", req.Name)
		assert.Equal(t, models.SecretTypeCredentials, req.Type)

		secretID := uuid.New()
		response := models.SecretResponse{
			ID:   secretID,
			Name: "Test Secret",
			Type: models.SecretTypeCredentials,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")
	client.SetMasterPassword("master-password")

	request := &models.SecretRequest{
		Name: "Test Secret",
		Type: models.SecretTypeCredentials,
		Data: map[string]any{
			"username": "user",
			"password": "pass",
		},
	}

	ctx := context.Background()
	response, err := client.CreateSecret(ctx, request)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, response.ID)
	assert.Equal(t, "Test Secret", response.Name)
}

func TestClient_GetSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/secrets", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "master-password", r.Header.Get("X-Master-Password"))

		secret1ID := uuid.New()
		secret2ID := uuid.New()
		response := models.SecretsListResponse{
			Secrets: []models.SecretResponse{
				{
					ID:   secret1ID,
					Name: "Secret 1",
					Type: models.SecretTypeCredentials,
				},
				{
					ID:   secret2ID,
					Name: "Secret 2",
					Type: models.SecretTypeText,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")
	client.SetMasterPassword("master-password")

	ctx := context.Background()
	response, err := client.GetSecrets(ctx)
	require.NoError(t, err)
	assert.Len(t, response.Secrets, 2)
	assert.NotEqual(t, uuid.Nil, response.Secrets[0].ID)
	assert.NotEqual(t, uuid.Nil, response.Secrets[1].ID)
}

func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	_, err := client.Register(ctx, "", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server error: Invalid request")
}

func TestClient_NetworkError(t *testing.T) {
	// Используем недоступный URL
	client := NewClient("http://localhost:99999")
	ctx := context.Background()

	_, err := client.Register(ctx, "user", "email", "pass")
	require.Error(t, err)
	// Проверяем что это сетевая ошибка (может быть разные сообщения в зависимости от ОС)
	assert.True(t, 
		strings.Contains(err.Error(), "connection refused") || 
		strings.Contains(err.Error(), "invalid port") ||
		strings.Contains(err.Error(), "dial tcp"),
		"Expected network error, got: %v", err)
}

func TestClient_DoRequest_MissingToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем что Authorization header отсутствует
		assert.Empty(t, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	// Не устанавливаем токен

	ctx := context.Background()
	request := &models.SecretRequest{
		Name: "Test",
		Type: models.SecretTypeText,
	}

	_, err := client.CreateSecret(ctx, request)
	require.Error(t, err)
}

func TestClient_DoRequest_MissingMasterPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		// Проверяем что X-Master-Password header отсутствует
		assert.Empty(t, r.Header.Get("X-Master-Password"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "ok"})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token")
	// Не устанавливаем мастер-пароль

	ctx := context.Background()
	request := &models.SecretRequest{
		Name: "Test",
		Type: models.SecretTypeText,
	}

	// Запрос должен пройти, но без заголовка X-Master-Password
	_, err := client.CreateSecret(ctx, request)
	require.NoError(t, err)
}
