package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
	"github.com/uryumtsevaa/gophkeeper/internal/models"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
)


// DataDependencies группирует зависимости для работы с данными
type DataDependencies struct {
	Repository  interfaces.Repository
	TxManager   interfaces.TransactionManager
}

// SecurityDependencies группирует зависимости для безопасности
type SecurityDependencies struct {
	Auth      interfaces.AuthServiceInterface
	CryptoSvc interfaces.CryptoService
}

// BusinessDependencies группирует зависимости для бизнес-логики
type BusinessDependencies struct {
	SyncSvc interfaces.SyncService
}

// ServiceDependencies группирует все зависимости сервиса
type ServiceDependencies struct {
	Data     DataDependencies
	Security SecurityDependencies
	Business BusinessDependencies
}

// AuthDomainService интерфейс для доменного сервиса аутентификации
type AuthDomainService interface {
	RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error)
	LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
}

// authDomainService реализация доменного сервиса аутентификации
type authDomainService struct {
	repo interfaces.Repository
	auth interfaces.AuthServiceInterface
}

// NewAuthDomainService создает новый доменный сервис аутентификации
func NewAuthDomainService(repo interfaces.Repository, auth interfaces.AuthServiceInterface) AuthDomainService {
	return &authDomainService{
		repo: repo,
		auth: auth,
	}
}

// RegisterUser регистрирует нового пользователя
func (a *authDomainService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	// Проверяем, не существует ли пользователь
	existingUser, err := a.repo.GetUserByUsername(ctx, req.Username)
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

	if err := a.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токен
	token, err := a.auth.GenerateToken(user.ID, user.Username)
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
func (a *authDomainService) LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Находим пользователя
	user, err := a.repo.GetUserByUsername(ctx, req.Username)
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
	token, err := a.auth.GenerateToken(user.ID, user.Username)
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

// SecretsDomainService интерфейс для доменного сервиса секретов
type SecretsDomainService interface {
	CreateSecret(ctx context.Context, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error)
	GetSecrets(ctx context.Context, userID uuid.UUID, masterPassword string) (*models.SecretsListResponse, error)
	GetSecret(ctx context.Context, secretID, userID uuid.UUID, masterPassword string) (*models.SecretResponse, error)
	UpdateSecret(ctx context.Context, secretID, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error)
	DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error
}

// secretsDomainService реализация доменного сервиса секретов
type secretsDomainService struct {
	repo      interfaces.Repository
	cryptoSvc interfaces.CryptoService
	txManager interfaces.TransactionManager
	syncSvc   interfaces.SyncService
}

// NewSecretsDomainService создает новый доменный сервис секретов
func NewSecretsDomainService(repo interfaces.Repository, cryptoSvc interfaces.CryptoService, txManager interfaces.TransactionManager, syncSvc interfaces.SyncService) SecretsDomainService {
	return &secretsDomainService{
		repo:      repo,
		cryptoSvc: cryptoSvc,
		txManager: txManager,
		syncSvc:   syncSvc,
	}
}

// CreateSecret создает новый секрет
func (s *secretsDomainService) CreateSecret(ctx context.Context, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
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
			fmt.Printf("Warning: failed to create sync operation: %v\n", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return buildSecretResponse(secret, req.Data), nil
}

// GetSecrets получает все секреты пользователя
func (s *secretsDomainService) GetSecrets(ctx context.Context, userID uuid.UUID, masterPassword string) (*models.SecretsListResponse, error) {
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

		responses = append(responses, *buildSecretResponse(secret, decryptedData))
	}

	return &models.SecretsListResponse{
		Secrets: responses,
		Total:   len(responses),
	}, nil
}

// GetSecret получает конкретный секрет
func (s *secretsDomainService) GetSecret(ctx context.Context, secretID, userID uuid.UUID, masterPassword string) (*models.SecretResponse, error) {
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

	return buildSecretResponse(secret, decryptedData), nil
}

// UpdateSecret обновляет секрет
func (s *secretsDomainService) UpdateSecret(ctx context.Context, secretID, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
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

	return buildSecretResponse(existing, req.Data), nil
}

// DeleteSecret удаляет секрет
func (s *secretsDomainService) DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error {
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

// buildSecretResponse создает ответ с секретом
func buildSecretResponse(secret *models.Secret, data any) *models.SecretResponse {
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

// SyncDomainService интерфейс для доменного сервиса синхронизации
type SyncDomainService interface {
	SyncSecrets(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error)
}

// syncDomainService реализация доменного сервиса синхронизации
type syncDomainService struct {
	syncSvc interfaces.SyncService
}

// NewSyncDomainService создает новый доменный сервис синхронизации
func NewSyncDomainService(syncSvc interfaces.SyncService) SyncDomainService {
	return &syncDomainService{
		syncSvc: syncSvc,
	}
}

// SyncSecrets синхронизирует секреты
func (s *syncDomainService) SyncSecrets(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error) {
	return s.syncSvc.ProcessSyncRequest(ctx, userID, req, masterPassword)
}

// DomainServices группирует доменные сервисы
type DomainServices struct {
	Auth    AuthDomainService
	Secrets SecretsDomainService
	Sync    SyncDomainService
}

// Service сервис для бизнес-логики
type Service struct {
	domains DomainServices
}

// NewService создает новый сервис
func NewService(domains DomainServices) *Service {
	return &Service{
		domains: domains,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *Service) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.LoginResponse, error) {
	return s.domains.Auth.RegisterUser(ctx, req)
}

// LoginUser авторизует пользователя
func (s *Service) LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	return s.domains.Auth.LoginUser(ctx, req)
}

// CreateSecret создает новый секрет
func (s *Service) CreateSecret(ctx context.Context, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
	return s.domains.Secrets.CreateSecret(ctx, userID, req, masterPassword)
}

// GetSecrets получает все секреты пользователя
func (s *Service) GetSecrets(ctx context.Context, userID uuid.UUID, masterPassword string) (*models.SecretsListResponse, error) {
	return s.domains.Secrets.GetSecrets(ctx, userID, masterPassword)
}

// GetSecret получает конкретный секрет
func (s *Service) GetSecret(ctx context.Context, secretID, userID uuid.UUID, masterPassword string) (*models.SecretResponse, error) {
	return s.domains.Secrets.GetSecret(ctx, secretID, userID, masterPassword)
}

// UpdateSecret обновляет секрет
func (s *Service) UpdateSecret(ctx context.Context, secretID, userID uuid.UUID, req *models.SecretRequest, masterPassword string) (*models.SecretResponse, error) {
	return s.domains.Secrets.UpdateSecret(ctx, secretID, userID, req, masterPassword)
}

// DeleteSecret удаляет секрет
func (s *Service) DeleteSecret(ctx context.Context, secretID, userID uuid.UUID) error {
	return s.domains.Secrets.DeleteSecret(ctx, secretID, userID)
}

// SyncSecrets синхронизирует секреты
func (s *Service) SyncSecrets(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error) {
	return s.domains.Sync.SyncSecrets(ctx, userID, req, masterPassword)
}