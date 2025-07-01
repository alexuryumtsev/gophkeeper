package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
)

// Client интерфейс для работы с GophKeeper API
type Client interface {
	// Authentication methods
	Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	
	// Token management
	SetToken(token string)
	GetToken() string
	SetMasterPassword(password string)
	GetMasterPassword() string
	
	// Secret management
	CreateSecret(ctx context.Context, req *SecretRequest) (*SecretResponse, error)
	GetSecrets(ctx context.Context) (*SecretsListResponse, error)
	GetSecret(ctx context.Context, secretID uuid.UUID) (*SecretResponse, error)
	UpdateSecret(ctx context.Context, secretID uuid.UUID, req *SecretRequest) (*SecretResponse, error)
	DeleteSecret(ctx context.Context, secretID uuid.UUID) error
	
	// Synchronization
	SyncSecrets(ctx context.Context, req *SyncRequest) (*SyncResponse, error)
	
	// Utility methods
	SetBaseURL(baseURL string) error
	Close() error
}

// client реализация интерфейса Client
type client struct {
	baseURL        *url.URL
	httpClient     *http.Client
	token          string
	masterPassword string
	userAgent      string
}

// NewClient создает новый клиент с заданной конфигурацией
func NewClient(config *Config) (Client, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	
	return &client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		userAgent: config.UserAgent,
	}, nil
}

// NewDefaultClient создает клиент с настройками по умолчанию
func NewDefaultClient(baseURL string) (Client, error) {
	config := DefaultConfig()
	config.BaseURL = baseURL
	return NewClient(config)
}

// Authentication methods

func (c *client) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	var resp LoginResponse
	err := c.doRequest(ctx, http.MethodPost, "/auth/register", req, &resp)
	if err != nil {
		return nil, err
	}
	
	c.SetToken(resp.Token)
	return &resp, nil
}

func (c *client) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, NewValidationError("credentials", "Username and password are required")
	}
	
	var resp LoginResponse
	err := c.doRequest(ctx, http.MethodPost, "/auth/login", req, &resp)
	if err != nil {
		return nil, err
	}
	
	c.SetToken(resp.Token)
	return &resp, nil
}

// Token management

func (c *client) SetToken(token string) {
	c.token = token
}

func (c *client) GetToken() string {
	return c.token
}

func (c *client) SetMasterPassword(password string) {
	c.masterPassword = password
}

func (c *client) GetMasterPassword() string {
	return c.masterPassword
}

// Secret management

func (c *client) CreateSecret(ctx context.Context, req *SecretRequest) (*SecretResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	var resp SecretResponse
	err := c.doRequestWithAuth(ctx, http.MethodPost, "/secrets", req, &resp)
	return &resp, err
}

func (c *client) GetSecrets(ctx context.Context) (*SecretsListResponse, error) {
	var resp SecretsListResponse
	err := c.doRequestWithAuth(ctx, http.MethodGet, "/secrets", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) GetSecret(ctx context.Context, secretID uuid.UUID) (*SecretResponse, error) {
	endpoint := fmt.Sprintf("/secrets/%s", secretID.String())
	var resp SecretResponse
	err := c.doRequestWithAuth(ctx, http.MethodGet, endpoint, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *client) UpdateSecret(ctx context.Context, secretID uuid.UUID, req *SecretRequest) (*SecretResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	
	endpoint := fmt.Sprintf("/secrets/%s", secretID.String())
	var resp SecretResponse
	err := c.doRequestWithAuth(ctx, http.MethodPut, endpoint, req, &resp)
	return &resp, err
}

func (c *client) DeleteSecret(ctx context.Context, secretID uuid.UUID) error {
	endpoint := fmt.Sprintf("/secrets/%s", secretID.String())
	return c.doRequestWithAuth(ctx, http.MethodDelete, endpoint, nil, nil)
}

// Synchronization

func (c *client) SyncSecrets(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	var resp SyncResponse
	err := c.doRequestWithAuth(ctx, http.MethodPost, "/sync", req, &resp)
	return &resp, err
}

// Utility methods

func (c *client) SetBaseURL(baseURL string) error {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	c.baseURL = parsedURL
	return nil
}

func (c *client) Close() error {
	// Закрываем соединения, если необходимо
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// Internal methods

func (c *client) doRequest(ctx context.Context, method, endpoint string, body, response any) error {
	return c.doRequestInternal(ctx, method, endpoint, body, response, false)
}

func (c *client) doRequestWithAuth(ctx context.Context, method, endpoint string, body, response any) error {
	return c.doRequestInternal(ctx, method, endpoint, body, response, true)
}

func (c *client) doRequestInternal(ctx context.Context, method, endpoint string, body, response any, requireAuth bool) error {
	// Создаем URL
	u := *c.baseURL
	u.Path = path.Join(APIPrefix, endpoint)
	
	// Подготавливаем тело запроса
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	// Создаем запрос
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return NewNetworkError(method, u.String(), err)
	}
	
	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	
	if requireAuth {
		if c.token == "" {
			return ErrUnauthorized
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		
		if c.masterPassword != "" {
			req.Header.Set("X-Master-Password", c.masterPassword)
		}
	}
	
	// Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NewNetworkError(method, u.String(), err)
	}
	defer resp.Body.Close()
	
	// Читаем тело ответа
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NewNetworkError("read response", u.String(), err)
	}
	
	// Обрабатываем ошибки HTTP
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.StatusCode, respBody)
	}
	
	// Декодируем успешный ответ
	if response != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

func (c *client) handleErrorResponse(statusCode int, body []byte) error {
	// Пытаемся декодировать стандартный формат ошибки
	var errorResp struct {
		Error string `json:"error"`
		Field string `json:"field,omitempty"`
	}
	
	message := "Unknown error"
	if len(body) > 0 {
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			message = errorResp.Error
		} else {
			// Если не удалось декодировать JSON, используем тело ответа как сообщение
			if strings.Contains(string(body), "html") {
				message = "Server returned HTML response"
			} else {
				message = string(body)
			}
		}
	}
	
	// Создаем соответствующую ошибку
	switch statusCode {
	case http.StatusUnauthorized:
		return NewAuthError(message)
	case http.StatusForbidden:
		return NewAPIError(statusCode, message)
	case http.StatusNotFound:
		return NewAPIError(statusCode, "Resource not found")
	case http.StatusConflict:
		return NewAPIError(statusCode, "Resource already exists")
	case http.StatusBadRequest:
		if errorResp.Field != "" {
			return NewValidationError(errorResp.Field, message)
		}
		return NewAPIError(statusCode, message)
	default:
		return NewAPIError(statusCode, message)
	}
}