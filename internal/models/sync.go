package models

import (
	"time"

	"github.com/google/uuid"
)

// SyncRequest запрос на синхронизацию
type SyncRequest struct {
	LastSyncTime time.Time            `json:"last_sync_time"`
	ClientHashes map[uuid.UUID]string `json:"client_hashes"`
}

// SyncResponse ответ на синхронизацию
type SyncResponse struct {
	UpdatedSecrets []SecretResponse `json:"updated_secrets"`
	DeletedSecrets []uuid.UUID      `json:"deleted_secrets"`
	SyncTime       time.Time        `json:"sync_time"`
}

// SyncOperation операция синхронизации
type SyncOperation struct {
	ID        uuid.UUID     `json:"id" db:"id"`
	UserID    uuid.UUID     `json:"user_id" db:"user_id"`
	SecretID  uuid.UUID     `json:"secret_id" db:"secret_id"`
	Operation OperationType `json:"operation" db:"operation"`
	Timestamp time.Time     `json:"timestamp" db:"timestamp"`
}

// OperationType тип операции
type OperationType string

const (
	OperationCreate OperationType = "create"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
)
