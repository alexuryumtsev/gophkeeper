package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// HTTPClientInterface интерфейс для HTTP клиента
type HTTPClientInterface interface {
	SetToken(token string)
	SetMasterPassword(password string)
	Register(ctx context.Context, username, email, password string) (*models.LoginResponse, error)
	Login(ctx context.Context, username, password string) (*models.LoginResponse, error)
	CreateSecret(ctx context.Context, req *models.SecretRequest) (*models.SecretResponse, error)
	GetSecrets(ctx context.Context) (*models.SecretsListResponse, error)
	GetSecret(ctx context.Context, secretID string) (*models.SecretResponse, error)
	UpdateSecret(ctx context.Context, secretID string, req *models.SecretRequest) (*models.SecretResponse, error)
	DeleteSecret(ctx context.Context, secretID string) error
	SyncSecrets(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error)
}

// Client HTTP клиент для взаимодействия с сервером
type Client struct {
	baseURL        string
	httpClient     *http.Client
	token          string
	masterPassword string
}

// NewClient создает новый клиент
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken устанавливает токен авторизации
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetMasterPassword устанавливает мастер-пароль
func (c *Client) SetMasterPassword(password string) {
	c.masterPassword = password
}

// doRequest выполняет HTTP запрос
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body any, response any) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.masterPassword != "" {
		req.Header.Set("X-Master-Password", c.masterPassword)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errorResp)
		if msg, ok := errorResp["error"].(string); ok {
			return fmt.Errorf("server error: %s", msg)
		}
		return fmt.Errorf("server error: status %d", resp.StatusCode)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, username, email, password string) (*models.LoginResponse, error) {
	req := models.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	var resp models.LoginResponse
	err := c.doRequest(ctx, "POST", "/api/v1/auth/register", req, &resp)
	if err != nil {
		return nil, err
	}

	c.SetToken(resp.Token)
	return &resp, nil
}

// Login авторизует пользователя
func (c *Client) Login(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	req := models.LoginRequest{
		Username: username,
		Password: password,
	}

	var resp models.LoginResponse
	err := c.doRequest(ctx, "POST", "/api/v1/auth/login", req, &resp)
	if err != nil {
		return nil, err
	}

	c.SetToken(resp.Token)
	return &resp, nil
}

// CreateSecret создает новый секрет
func (c *Client) CreateSecret(ctx context.Context, req *models.SecretRequest) (*models.SecretResponse, error) {
	var resp models.SecretResponse
	err := c.doRequest(ctx, "POST", "/api/v1/secrets", req, &resp)
	return &resp, err
}

// GetSecrets получает список секретов
func (c *Client) GetSecrets(ctx context.Context) (*models.SecretsListResponse, error) {
	var resp models.SecretsListResponse
	err := c.doRequest(ctx, "GET", "/api/v1/secrets", nil, &resp)
	return &resp, err
}

// GetSecret получает конкретный секрет
func (c *Client) GetSecret(ctx context.Context, secretID string) (*models.SecretResponse, error) {
	var resp models.SecretResponse
	err := c.doRequest(ctx, "GET", "/api/v1/secrets/"+secretID, nil, &resp)
	return &resp, err
}

// UpdateSecret обновляет секрет
func (c *Client) UpdateSecret(ctx context.Context, secretID string, req *models.SecretRequest) (*models.SecretResponse, error) {
	var resp models.SecretResponse
	err := c.doRequest(ctx, "PUT", "/api/v1/secrets/"+secretID, req, &resp)
	return &resp, err
}

// DeleteSecret удаляет секрет
func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
	return c.doRequest(ctx, "DELETE", "/api/v1/secrets/"+secretID, nil, nil)
}

// SyncSecrets синхронизирует секреты
func (c *Client) SyncSecrets(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error) {
	var resp models.SyncResponse
	err := c.doRequest(ctx, "POST", "/api/v1/sync", req, &resp)
	return &resp, err
}
