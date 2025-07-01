package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

type SyncService interface {
	CreateSyncOperation(ctx context.Context, userID, secretID uuid.UUID, operation models.OperationType) error
	ProcessSyncRequest(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error)
}

type syncService struct {
	repo          Repository
	cryptoService CryptoService
}

func NewSyncService(repo Repository, cryptoService CryptoService) SyncService {
	return &syncService{
		repo:          repo,
		cryptoService: cryptoService,
	}
}

func (s *syncService) CreateSyncOperation(ctx context.Context, userID, secretID uuid.UUID, operation models.OperationType) error {
	syncOp := &models.SyncOperation{
		ID:        uuid.New(),
		UserID:    userID,
		SecretID:  secretID,
		Operation: operation,
		Timestamp: time.Now(),
	}

	if err := s.repo.CreateSyncOperation(ctx, syncOp); err != nil {
		return fmt.Errorf("failed to create sync operation: %w", err)
	}

	return nil
}

func (s *syncService) ProcessSyncRequest(ctx context.Context, userID uuid.UUID, req *models.SyncRequest, masterPassword string) (*models.SyncResponse, error) {
	syncOps, err := s.repo.GetSyncOperationsAfter(ctx, userID, req.LastSyncTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync operations: %w", err)
	}

	updatedSecretIDs := make(map[uuid.UUID]bool)
	deletedSecrets := make([]uuid.UUID, 0)

	for _, op := range syncOps {
		switch op.Operation {
		case models.OperationCreate, models.OperationUpdate:
			updatedSecretIDs[op.SecretID] = true
		case models.OperationDelete:
			deletedSecrets = append(deletedSecrets, op.SecretID)
			delete(updatedSecretIDs, op.SecretID)
		}
	}

	updatedSecrets := make([]models.SecretResponse, 0)
	for secretID := range updatedSecretIDs {
		secret, err := s.repo.GetSecretByID(ctx, secretID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get secret %s: %w", secretID, err)
		}
		if secret == nil {
			continue
		}

		decryptedData, err := s.cryptoService.DecryptSecretData(secret.Data, masterPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret data: %w", err)
		}

		secretResponse := models.SecretResponse{
			ID:        secret.ID,
			Name:      secret.Name,
			Type:      secret.Type,
			Data:      decryptedData,
			Metadata:  secret.Metadata,
			SyncHash:  secret.SyncHash,
			CreatedAt: secret.CreatedAt,
			UpdatedAt: secret.UpdatedAt,
		}
		updatedSecrets = append(updatedSecrets, secretResponse)
	}

	response := &models.SyncResponse{
		UpdatedSecrets: updatedSecrets,
		DeletedSecrets: deletedSecrets,
		SyncTime:       time.Now(),
	}

	return response, nil
}