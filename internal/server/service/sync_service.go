package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
	"github.com/uryumtsevaa/gophkeeper/internal/server/interfaces"
)

// syncService реализация сервиса синхронизации
type syncService struct {
	repo          interfaces.Repository
	cryptoService interfaces.CryptoService
}

// NewSyncService создает новый сервис синхронизации
func NewSyncService(repo interfaces.Repository, cryptoService interfaces.CryptoService) interfaces.SyncService {
	return &syncService{
		repo:          repo,
		cryptoService: cryptoService,
	}
}

// CreateSyncOperation создает операцию синхронизации
func (s *syncService) CreateSyncOperation(ctx context.Context, userID, secretID uuid.UUID, operation models.OperationType) error {
	syncOp := &models.SyncOperation{
		ID:        uuid.New(),
		UserID:    userID,
		SecretID:  secretID,
		Operation: operation,
		Timestamp: time.Now(),
	}
	return s.repo.CreateSyncOperation(ctx, syncOp)
}

// ProcessSyncRequest обрабатывает запрос синхронизации
func (s *syncService) ProcessSyncRequest(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error) {
	// Получаем секреты, измененные после указанной даты
	secrets, err := s.repo.GetSecretsModifiedAfter(ctx, userID, req.LastSyncTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get modified secrets: %w", err)
	}

	var secretResponses []models.SecretResponse
	for _, secret := range secrets {
		// Расшифровываем данные для передачи
		decryptedData, err := s.cryptoService.DecryptSecretData(secret.Data, masterPassword)
		if err != nil {
			continue // Пропускаем секреты, которые не удается расшифровать
		}

		secretResponses = append(secretResponses, models.SecretResponse{
			ID:        secret.ID,
			Type:      secret.Type,
			Name:      secret.Name,
			Data:      decryptedData,
			Metadata:  secret.Metadata,
			CreatedAt: secret.CreatedAt,
			UpdatedAt: secret.UpdatedAt,
			SyncHash:  secret.SyncHash,
		})
	}

	return &models.SyncResponse{
		UpdatedSecrets: secretResponses,
		DeletedSecrets: []uuid.UUID{},
		SyncTime:       time.Now(),
	}, nil
}