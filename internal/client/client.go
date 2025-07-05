package client

import (
	"context"

	"github.com/uryumtsevaa/gophkeeper/internal/client/middleware"
	"github.com/uryumtsevaa/gophkeeper/internal/client/router"
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
	router         router.Router
	token          string
	masterPassword string
}

// NewClient создает новый клиент
func NewClient(baseURL string) *Client {
	httpRouter := router.NewHTTPRouter(baseURL)
	return &Client{
		router: httpRouter,
	}
}

// SetToken устанавливает токен авторизации
func (c *Client) SetToken(token string) {
	c.token = token
	c.setupMiddleware()
}

// SetMasterPassword устанавливает мастер-пароль
func (c *Client) SetMasterPassword(password string) {
	c.masterPassword = password
	c.setupMiddleware()
}

// setupMiddleware настраивает middleware для роутера
func (c *Client) setupMiddleware() {
	if httpRouter, ok := c.router.(*router.HTTPRouter); ok {
		httpRouter.Use(middleware.AuthMiddleware(c.token))
		httpRouter.Use(middleware.MasterPasswordMiddleware(c.masterPassword))
	}
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, username, email, password string) (*models.LoginResponse, error) {
	resp, err := c.router.Register(ctx, username, email, password)
	if err != nil {
		return nil, err
	}

	c.SetToken(resp.Token)
	return resp, nil
}

// Login авторизует пользователя
func (c *Client) Login(ctx context.Context, username, password string) (*models.LoginResponse, error) {
	resp, err := c.router.Login(ctx, username, password)
	if err != nil {
		return nil, err
	}

	c.SetToken(resp.Token)
	return resp, nil
}

// CreateSecret создает новый секрет
func (c *Client) CreateSecret(ctx context.Context, req *models.SecretRequest) (*models.SecretResponse, error) {
	return c.router.CreateSecret(ctx, req)
}

// GetSecrets получает список секретов
func (c *Client) GetSecrets(ctx context.Context) (*models.SecretsListResponse, error) {
	return c.router.GetSecrets(ctx)
}

// GetSecret получает конкретный секрет
func (c *Client) GetSecret(ctx context.Context, secretID string) (*models.SecretResponse, error) {
	return c.router.GetSecret(ctx, secretID)
}

// UpdateSecret обновляет секрет
func (c *Client) UpdateSecret(ctx context.Context, secretID string, req *models.SecretRequest) (*models.SecretResponse, error) {
	return c.router.UpdateSecret(ctx, secretID, req)
}

// DeleteSecret удаляет секрет
func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
	return c.router.DeleteSecret(ctx, secretID)
}

// SyncSecrets синхронизирует секреты
func (c *Client) SyncSecrets(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error) {
	return c.router.SyncSecrets(ctx, req)
}
