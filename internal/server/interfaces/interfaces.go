package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// ServiceInterface интерфейс для бизнес-логики
type ServiceInterface interface {
	RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error)
	LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	CreateSecret(ctx context.Context, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error)
	GetSecrets(ctx context.Context, userID uuid.UUID, masterPassword string) (*models.SecretsListResponse, error)
	GetSecret(ctx context.Context, secretID, userID uuid.UUID, masterPassword string) (*models.SecretResponse, error)
	UpdateSecret(ctx context.Context, secretID, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error)
	DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error
	SyncSecrets(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error)
}

// Repository интерфейс для работы с данными
type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	CreateSecret(ctx context.Context, secret *models.Secret) error
	GetSecretsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Secret, error)
	GetSecretByID(ctx context.Context, secretID, userID uuid.UUID) (*models.Secret, error)
	UpdateSecret(ctx context.Context, secret *models.Secret) error
	DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error
	GetSecretsModifiedAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.Secret, error)
	CreateSyncOperation(ctx context.Context, operation *models.SyncOperation) error
	GetSyncOperationsAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.SyncOperation, error)
}

// AuthServiceInterface интерфейс для сервиса аутентификации
type AuthServiceInterface interface {
	GenerateToken(userID uuid.UUID, username string) (string, error)
	ValidateToken(token string) (interface{}, error)
}

// CryptoService интерфейс для сервиса шифрования
type CryptoService interface {
	EncryptSecretData(data interface{}, masterPassword string) ([]byte, error)
	DecryptSecretData(encryptedData []byte, masterPassword string) (interface{}, error)
}

// TransactionManager интерфейс для управления транзакциями
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
}

// SyncService интерфейс для сервиса синхронизации
type SyncService interface {
	CreateSyncOperation(ctx context.Context, userID, secretID uuid.UUID, operation models.OperationType) error
	ProcessSyncRequest(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error)
}

// HandlerInterface интерфейс для обработчиков HTTP запросов  
type HandlerInterface interface {
	Register() func(c interface{})
	Login() func(c interface{})
	CreateSecret() func(c interface{})
	GetSecrets() func(c interface{})
	GetSecret() func(c interface{})
	UpdateSecret() func(c interface{})
	DeleteSecret() func(c interface{})
	SyncSecrets() func(c interface{})
}