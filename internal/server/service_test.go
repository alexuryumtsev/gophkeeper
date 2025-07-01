package server

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// MockRepository для тестирования
type MockRepository struct {
	users       map[string]*models.User
	secrets     map[uuid.UUID]*models.Secret
	syncOps     []*models.SyncOperation
	userSecrets map[uuid.UUID][]*models.Secret
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:       make(map[string]*models.User),
		secrets:     make(map[uuid.UUID]*models.Secret),
		syncOps:     make([]*models.SyncOperation, 0),
		userSecrets: make(map[uuid.UUID][]*models.Secret),
	}
}

func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
	m.users[user.Username] = user
	return nil
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, exists := m.users[username]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) CreateSecret(ctx context.Context, secret *models.Secret) error {
	m.secrets[secret.ID] = secret
	m.userSecrets[secret.UserID] = append(m.userSecrets[secret.UserID], secret)
	return nil
}

func (m *MockRepository) GetSecretsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Secret, error) {
	return m.userSecrets[userID], nil
}

func (m *MockRepository) GetSecretByID(ctx context.Context, secretID uuid.UUID, userID uuid.UUID) (*models.Secret, error) {
	secret, exists := m.secrets[secretID]
	if !exists || secret.UserID != userID {
		return nil, nil
	}
	return secret, nil
}

func (m *MockRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	existing, exists := m.secrets[secret.ID]
	if !exists || existing.UserID != secret.UserID {
		return nil
	}
	m.secrets[secret.ID] = secret
	return nil
}

func (m *MockRepository) DeleteSecret(ctx context.Context, secretID uuid.UUID, userID uuid.UUID) error {
	secret, exists := m.secrets[secretID]
	if !exists || secret.UserID != userID {
		return nil
	}
	delete(m.secrets, secretID)

	// Удаляем из userSecrets
	userSecrets := m.userSecrets[userID]
	for i, s := range userSecrets {
		if s.ID == secretID {
			m.userSecrets[userID] = append(userSecrets[:i], userSecrets[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockRepository) GetSecretsModifiedAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.Secret, error) {
	var result []*models.Secret
	for _, secret := range m.userSecrets[userID] {
		if secret.UpdatedAt.After(after) {
			result = append(result, secret)
		}
	}
	return result, nil
}

func (m *MockRepository) CreateSyncOperation(ctx context.Context, operation *models.SyncOperation) error {
	m.syncOps = append(m.syncOps, operation)
	return nil
}

func (m *MockRepository) GetSyncOperationsAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.SyncOperation, error) {
	var result []*models.SyncOperation
	for _, op := range m.syncOps {
		if op.UserID == userID && op.Timestamp.After(after) {
			result = append(result, op)
		}
	}
	return result, nil
}

func TestService_RegisterUser(t *testing.T) {
	repo := NewMockRepository()
	auth := NewAuthService("test-secret")
	encryptor := crypto.NewAESEncryptor()
	cryptoSvc := NewCryptoService(encryptor)
	syncSvc := NewSyncService(repo, cryptoSvc)
	txManager := NewMockTransactionManager()
	service := NewService(repo, auth, cryptoSvc, syncSvc, txManager)

	req := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	ctx := context.Background()
	response, err := service.RegisterUser(ctx, req)
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	if response.User.Username != req.Username {
		t.Errorf("Expected username %s, got %s", req.Username, response.User.Username)
	}

	if response.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, response.User.Email)
	}

	if response.Token == "" {
		t.Error("Expected non-empty token")
	}

	// Проверяем, что пользователь сохранен в репозитории
	savedUser, err := repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		t.Fatalf("Failed to get saved user: %v", err)
	}

	if savedUser == nil {
		t.Fatal("User was not saved")
	}

	// Проверяем, что пароль захеширован
	if !crypto.VerifyPassword(req.Password, savedUser.PasswordHash) {
		t.Error("Password was not properly hashed")
	}
}

func TestService_LoginUser(t *testing.T) {
	repo := NewMockRepository()
	auth := NewAuthService("test-secret")
	encryptor := crypto.NewAESEncryptor()
	cryptoSvc := NewCryptoService(encryptor)
	syncSvc := NewSyncService(repo, cryptoSvc)
	txManager := NewMockTransactionManager()
	service := NewService(repo, auth, cryptoSvc, syncSvc, txManager)

	// Сначала регистрируем пользователя
	registerReq := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	ctx := context.Background()
	_, err := service.RegisterUser(ctx, registerReq)
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	// Теперь пытаемся войти
	loginReq := &models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	response, err := service.LoginUser(ctx, loginReq)
	if err != nil {
		t.Fatalf("LoginUser failed: %v", err)
	}

	if response.User.Username != loginReq.Username {
		t.Errorf("Expected username %s, got %s", loginReq.Username, response.User.Username)
	}

	if response.Token == "" {
		t.Error("Expected non-empty token")
	}

	// Тест с неверным паролем
	wrongPasswordReq := &models.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	_, err = service.LoginUser(ctx, wrongPasswordReq)
	if err == nil {
		t.Error("Expected login to fail with wrong password")
	}
}

func TestService_CreateSecret(t *testing.T) {
	repo := NewMockRepository()
	auth := NewAuthService("test-secret")
	encryptor := crypto.NewAESEncryptor()
	cryptoSvc := NewCryptoService(encryptor)
	syncSvc := NewSyncService(repo, cryptoSvc)
	txManager := NewMockTransactionManager()
	service := NewService(repo, auth, cryptoSvc, syncSvc, txManager)

	userID := uuid.New()
	masterPassword := "master-password"

	credentials := models.Credentials{
		Name:     "Test Credentials",
		Username: "user",
		Password: "pass",
		URL:      "https://example.com",
	}

	req := &models.SecretRequest{
		Type:     models.SecretTypeCredentials,
		Name:     "Test Secret",
		Data:     credentials,
		Metadata: "test metadata",
	}

	ctx := context.Background()
	response, err := service.CreateSecret(ctx, userID, req, masterPassword)
	if err != nil {
		t.Fatalf("CreateSecret failed: %v", err)
	}

	if response.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, response.Name)
	}

	if response.Type != req.Type {
		t.Errorf("Expected type %s, got %s", req.Type, response.Type)
	}

	// Проверяем, что секрет сохранен в репозитории
	savedSecret, err := repo.GetSecretByID(ctx, response.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get saved secret: %v", err)
	}

	if savedSecret == nil {
		t.Fatal("Secret was not saved")
	}

	// Проверяем, что данные зашифрованы
	if len(savedSecret.Data) == 0 {
		t.Error("Secret data should not be empty")
	}
}

func TestService_GetSecrets(t *testing.T) {
	repo := NewMockRepository()
	auth := NewAuthService("test-secret")
	encryptor := crypto.NewAESEncryptor()
	cryptoSvc := NewCryptoService(encryptor)
	syncSvc := NewSyncService(repo, cryptoSvc)
	txManager := NewMockTransactionManager()
	service := NewService(repo, auth, cryptoSvc, syncSvc, txManager)

	userID := uuid.New()
	masterPassword := "master-password"

	// Создаем несколько секретов
	for i := 0; i < 3; i++ {
		credentials := models.Credentials{
			Name:     "Test Credentials",
			Username: "user",
			Password: "pass",
		}

		req := &models.SecretRequest{
			Type: models.SecretTypeCredentials,
			Name: "Test Secret",
			Data: credentials,
		}

		ctx := context.Background()
		_, err := service.CreateSecret(ctx, userID, req, masterPassword)
		if err != nil {
			t.Fatalf("CreateSecret failed: %v", err)
		}
	}

	// Получаем список секретов
	ctx := context.Background()
	response, err := service.GetSecrets(ctx, userID, masterPassword)
	if err != nil {
		t.Fatalf("GetSecrets failed: %v", err)
	}

	if len(response.Secrets) != 3 {
		t.Errorf("Expected 3 secrets, got %d", len(response.Secrets))
	}

	if response.Total != 3 {
		t.Errorf("Expected total 3, got %d", response.Total)
	}
}

func TestAuthService_GenerateAndValidateToken(t *testing.T) {
	auth := NewAuthService("test-secret")
	userID := uuid.New()
	username := "testuser"

	// Генерируем токен
	token, err := auth.GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Валидируем токен
	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.Username != username {
		t.Errorf("Expected Username %s, got %s", username, claims.Username)
	}

	// Тест с невалидным токеном
	_, err = auth.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected validation to fail for invalid token")
	}
}
