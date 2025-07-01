package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
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

// Service сервис для бизнес-логики
type Service struct {
	repo         Repository
	auth         AuthServiceInterface
	cryptoSvc    CryptoService
	syncSvc      SyncService
	txManager    TransactionManager
}

// NewService создает новый сервис
func NewService(repo Repository, auth AuthServiceInterface, cryptoSvc CryptoService, syncSvc SyncService, txManager TransactionManager) *Service {
	return &Service{
		repo:      repo,
		auth:      auth,
		cryptoSvc: cryptoSvc,
		syncSvc:   syncSvc,
		txManager: txManager,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *Service) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	// Проверяем, не существует ли пользователь
	existingUser, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Создаем нового пользователя
	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: crypto.HashPassword(req.Password),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токен
	token, err := s.auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Создаем копию пользователя для ответа без пароля
	responseUser := *user
	responseUser.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  responseUser,
	}, nil
}

// LoginUser авторизует пользователя
func (s *Service) LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Находим пользователя
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Проверяем пароль
	if !crypto.VerifyPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Генерируем токен
	token, err := s.auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Создаем копию пользователя для ответа без пароля
	responseUser := *user
	responseUser.PasswordHash = ""

	return &models.LoginResponse{
		Token: token,
		User:  responseUser,
	}, nil
}

// CreateSecret создает новый секрет
func (s *Service) CreateSecret(ctx context.Context, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
	// Шифруем данные
	encryptedData, err := s.cryptoSvc.EncryptSecretData(req.Data, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Создаем секрет
	secret := &models.Secret{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      req.Type,
		Name:      req.Name,
		Metadata:  req.Metadata,
		Data:      encryptedData,
		SyncHash:  crypto.GenerateSyncHash(encryptedData),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Используем транзакцию для создания секрета и операции синхронизации
	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateSecret(txCtx, secret); err != nil {
			return fmt.Errorf("failed to create secret: %w", err)
		}

		// Создаем операцию синхронизации
		if err := s.syncSvc.CreateSyncOperation(txCtx, userID, secret.ID, models.OperationCreate); err != nil {
			// Логируем ошибку, но не прерываем операцию
			fmt.Printf("Warning: failed to create sync operation: %v\n", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.buildSecretResponse(secret, req.Data), nil
}

// GetSecrets получает все секреты пользователя
func (s *Service) GetSecrets(ctx context.Context, userID uuid.UUID, masterPassword string) (*models.SecretsListResponse, error) {
	secrets, err := s.repo.GetSecretsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	var responses []models.SecretResponse
	for _, secret := range secrets {
		// Расшифровываем данные
		decryptedData, err := s.cryptoSvc.DecryptSecretData(secret.Data, masterPassword)
		if err != nil {
			// Пропускаем секреты, которые не удается расшифровать
			continue
		}

		responses = append(responses, *s.buildSecretResponse(secret, decryptedData))
	}

	return &models.SecretsListResponse{
		Secrets: responses,
		Total:   len(responses),
	}, nil
}

// GetSecret получает конкретный секрет
func (s *Service) GetSecret(ctx context.Context, secretID, userID uuid.UUID, masterPassword string) (*models.SecretResponse, error) {
	secret, err := s.repo.GetSecretByID(ctx, secretID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Расшифровываем данные
	decryptedData, err := s.cryptoSvc.DecryptSecretData(secret.Data, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}

	return s.buildSecretResponse(secret, decryptedData), nil
}

// UpdateSecret обновляет секрет
func (s *Service) UpdateSecret(ctx context.Context, secretID, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
	// Проверяем, что секрет существует и принадлежит пользователю
	existing, err := s.repo.GetSecretByID(ctx, secretID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("secret not found")
	}

	// Шифруем новые данные
	encryptedData, err := s.cryptoSvc.EncryptSecretData(req.Data, masterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Обновляем секрет
	existing.Name = req.Name
	existing.Metadata = req.Metadata
	existing.Data = encryptedData
	existing.SyncHash = crypto.GenerateSyncHash(encryptedData)
	existing.UpdatedAt = time.Now()

	// Используем транзакцию для обновления секрета и создания операции синхронизации
	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.UpdateSecret(txCtx, existing); err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}

		// Создаем операцию синхронизации
		if err := s.syncSvc.CreateSyncOperation(txCtx, userID, secretID, models.OperationUpdate); err != nil {
			fmt.Printf("Warning: failed to create sync operation: %v\n", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.buildSecretResponse(existing, req.Data), nil
}

// DeleteSecret удаляет секрет
func (s *Service) DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error {
	// Используем транзакцию для удаления секрета и создания операции синхронизации
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.DeleteSecret(txCtx, secretID, userID); err != nil {
			return fmt.Errorf("failed to delete secret: %w", err)
		}

		// Создаем операцию синхронизации
		if err := s.syncSvc.CreateSyncOperation(txCtx, userID, secretID, models.OperationDelete); err != nil {
			fmt.Printf("Warning: failed to create sync operation: %v\n", err)
		}

		return nil
	})
}

// SyncSecrets синхронизирует секреты
func (s *Service) SyncSecrets(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error) {
	return s.syncSvc.ProcessSyncRequest(ctx, userID, req, masterPassword)
}

// buildSecretResponse создает ответ с секретом
func (s *Service) buildSecretResponse(secret *models.Secret, data interface{}) *models.SecretResponse {
	return &models.SecretResponse{
		ID:        secret.ID,
		Type:      secret.Type,
		Name:      secret.Name,
		Data:      data,
		Metadata:  secret.Metadata,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
		SyncHash:  secret.SyncHash,
	}
}