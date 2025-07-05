package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/uryumtsevaa/gophkeeper/internal/crypto"
	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// Repository интерфейс для работы с данными
type Repository interface {
	// Users
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)

	// Secrets
	CreateSecret(ctx context.Context, secret *models.Secret) error
	GetSecretsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Secret, error)
	GetSecretByID(ctx context.Context, secretID uuid.UUID, userID uuid.UUID) (*models.Secret, error)
	UpdateSecret(ctx context.Context, secret *models.Secret) error
	DeleteSecret(ctx context.Context, secretID uuid.UUID, userID uuid.UUID) error

	// Sync
	GetSecretsModifiedAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.Secret, error)
	CreateSyncOperation(ctx context.Context, operation *models.SyncOperation) error
	GetSyncOperationsAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.SyncOperation, error)
}

// DBExecutor интерфейс для выполнения SQL операций
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db        DBExecutor
	encryptor crypto.Encryptor
}

// NewPostgresRepository создает новый PostgreSQL репозиторий
func NewPostgresRepository(db DBExecutor, encryptor crypto.Encryptor) *PostgresRepository {
	return &PostgresRepository{
		db:        db,
		encryptor: encryptor,
	}
}

// CreateUser создает нового пользователя
func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByUsername получает пользователя по имени
func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE username = $1`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// CreateSecret создает новый секрет
func (r *PostgresRepository) CreateSecret(ctx context.Context, secret *models.Secret) error {
	query := `
		INSERT INTO secrets (id, user_id, type, name, metadata, data, sync_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(ctx, query,
		secret.ID,
		secret.UserID,
		secret.Type,
		secret.Name,
		secret.Metadata,
		secret.Data,
		secret.SyncHash,
		secret.CreatedAt,
		secret.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetSecretsByUserID получает все секреты пользователя
func (r *PostgresRepository) GetSecretsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, type, name, metadata, data, sync_hash, created_at, updated_at
		FROM secrets WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		var secret models.Secret
		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Type,
			&secret.Name,
			&secret.Metadata,
			&secret.Data,
			&secret.SyncHash,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

// GetSecretByID получает секрет по ID
func (r *PostgresRepository) GetSecretByID(ctx context.Context, secretID uuid.UUID, userID uuid.UUID) (*models.Secret, error) {
	var secret models.Secret
	query := `
		SELECT id, user_id, type, name, metadata, data, sync_hash, created_at, updated_at
		FROM secrets WHERE id = $1 AND user_id = $2`

	err := r.db.QueryRow(ctx, query, secretID, userID).Scan(
		&secret.ID,
		&secret.UserID,
		&secret.Type,
		&secret.Name,
		&secret.Metadata,
		&secret.Data,
		&secret.SyncHash,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return &secret, nil
}

// UpdateSecret обновляет секрет
func (r *PostgresRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	query := `
		UPDATE secrets 
		SET name = $3, metadata = $4, data = $5, sync_hash = $6, updated_at = $7
		WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query,
		secret.ID,
		secret.UserID,
		secret.Name,
		secret.Metadata,
		secret.Data,
		secret.SyncHash,
		secret.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("secret not found or access denied")
	}

	return nil
}

// DeleteSecret удаляет секрет
func (r *PostgresRepository) DeleteSecret(ctx context.Context, secretID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM secrets WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, secretID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("secret not found or access denied")
	}

	return nil
}

// GetSecretsModifiedAfter получает секреты, измененные после указанного времени
func (r *PostgresRepository) GetSecretsModifiedAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, type, name, metadata, data, sync_hash, created_at, updated_at
		FROM secrets WHERE user_id = $1 AND updated_at > $2 ORDER BY updated_at`

	rows, err := r.db.Query(ctx, query, userID, after)
	if err != nil {
		return nil, fmt.Errorf("failed to get modified secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		var secret models.Secret
		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Type,
			&secret.Name,
			&secret.Metadata,
			&secret.Data,
			&secret.SyncHash,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}
		secrets = append(secrets, &secret)
	}

	return secrets, nil
}

// CreateSyncOperation создает операцию синхронизации
func (r *PostgresRepository) CreateSyncOperation(ctx context.Context, operation *models.SyncOperation) error {
	query := `
		INSERT INTO sync_operations (id, user_id, secret_id, operation, timestamp)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query,
		operation.ID,
		operation.UserID,
		operation.SecretID,
		operation.Operation,
		operation.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create sync operation: %w", err)
	}

	return nil
}

// GetSyncOperationsAfter получает операции синхронизации после указанного времени
func (r *PostgresRepository) GetSyncOperationsAfter(ctx context.Context, userID uuid.UUID, after time.Time) ([]*models.SyncOperation, error) {
	query := `
		SELECT id, user_id, secret_id, operation, timestamp
		FROM sync_operations WHERE user_id = $1 AND timestamp > $2 ORDER BY timestamp`

	rows, err := r.db.Query(ctx, query, userID, after)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync operations: %w", err)
	}
	defer rows.Close()

	var operations []*models.SyncOperation
	for rows.Next() {
		var operation models.SyncOperation
		var secretID sql.NullString

		err := rows.Scan(
			&operation.ID,
			&operation.UserID,
			&secretID,
			&operation.Operation,
			&operation.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sync operation: %w", err)
		}

		if secretID.Valid {
			if id, err := uuid.Parse(secretID.String); err == nil {
				operation.SecretID = id
			}
		}

		operations = append(operations, &operation)
	}

	return operations, nil
}
