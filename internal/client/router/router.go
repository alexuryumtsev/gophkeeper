package router

import (
	"context"
	"net/http"
	"time"

	"github.com/uryumtsevaa/gophkeeper/internal/client/common"
	httpclient "github.com/uryumtsevaa/gophkeeper/internal/client/http"
	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// Router интерфейс для роутера клиента
type Router interface {
	Register(ctx context.Context, username, email, password string) (*models.LoginResponse, error)
	Login(ctx context.Context, username, password string) (*models.LoginResponse, error)
	CreateSecret(ctx context.Context, req *models.SecretRequest) (*models.SecretResponse, error)
	GetSecrets(ctx context.Context) (*models.SecretsListResponse, error)
	GetSecret(ctx context.Context, secretID string) (*models.SecretResponse, error)
	UpdateSecret(ctx context.Context, secretID string, req *models.SecretRequest) (*models.SecretResponse, error)
	DeleteSecret(ctx context.Context, secretID string) error
	SyncSecrets(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error)
}

// HTTPRouter HTTP роутер для взаимодействия с API
type HTTPRouter struct {
	baseURL    string
	httpClient *http.Client
	middleware []common.Middleware
}

// NewHTTPRouter создает новый HTTP роутер
func NewHTTPRouter(baseURL string) *HTTPRouter {
	return &HTTPRouter{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		middleware: []common.Middleware{},
	}
}

// Use добавляет middleware в цепочку
func (r *HTTPRouter) Use(middleware common.Middleware) {
	r.middleware = append(r.middleware, middleware)
}

// DoRequest выполняет HTTP запрос с применением middleware
func (r *HTTPRouter) DoRequest(ctx context.Context, method, endpoint string, body any, response any) error {
	// Создаем базовый запрос
	req := &common.Request{
		Method:   method,
		Endpoint: endpoint,
		Body:     body,
		Response: response,
		Context:  ctx,
	}

	// Применяем middleware
	handler := r.executeRequest
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}

	return handler(req)
}

// executeRequest выполняет HTTP запрос
func (r *HTTPRouter) executeRequest(req *common.Request) error {
	return httpclient.DoRequest(r.httpClient, r.baseURL, req)
}

// Register регистрирует нового пользователя
func (r *HTTPRouter) Register(ctx context.Context, username, email, password string) (*models.LoginResponse, error) {
	req := models.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	var resp models.LoginResponse
	err := r.DoRequest(ctx, "POST", "/api/v1/auth/register", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Login авторизует пользователя
func (r *HTTPRouter) Login(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	req := models.LoginRequest{
		Username: username,
		Password: password,
	}

	var resp models.LoginResponse
	err := r.DoRequest(ctx, "POST", "/api/v1/auth/login", req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateSecret создает новый секрет
func (r *HTTPRouter) CreateSecret(ctx context.Context, req *models.SecretRequest) (*models.SecretResponse, error) {
	var resp models.SecretResponse
	err := r.DoRequest(ctx, "POST", "/api/v1/secrets", req, &resp)
	return &resp, err
}

// GetSecrets получает список секретов
func (r *HTTPRouter) GetSecrets(ctx context.Context) (*models.SecretsListResponse, error) {
	var resp models.SecretsListResponse
	err := r.DoRequest(ctx, "GET", "/api/v1/secrets", nil, &resp)
	return &resp, err
}

// GetSecret получает конкретный секрет
func (r *HTTPRouter) GetSecret(ctx context.Context, secretID string) (*models.SecretResponse, error) {
	var resp models.SecretResponse
	err := r.DoRequest(ctx, "GET", "/api/v1/secrets/"+secretID, nil, &resp)
	return &resp, err
}

// UpdateSecret обновляет секрет
func (r *HTTPRouter) UpdateSecret(ctx context.Context, secretID string, req *models.SecretRequest) (*models.SecretResponse, error) {
	var resp models.SecretResponse
	err := r.DoRequest(ctx, "PUT", "/api/v1/secrets/"+secretID, req, &resp)
	return &resp, err
}

// DeleteSecret удаляет секрет
func (r *HTTPRouter) DeleteSecret(ctx context.Context, secretID string) error {
	return r.DoRequest(ctx, "DELETE", "/api/v1/secrets/"+secretID, nil, nil)
}

// SyncSecrets синхронизирует секреты
func (r *HTTPRouter) SyncSecrets(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error) {
	var resp models.SyncResponse
	err := r.DoRequest(ctx, "POST", "/api/v1/sync", req, &resp)
	return &resp, err
}
