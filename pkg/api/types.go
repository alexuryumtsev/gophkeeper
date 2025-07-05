// Package api provides public types and interfaces for the GophKeeper API client.
package api

import (
	"time"

	"github.com/google/uuid"
)

// API версии и конфигурация
const (
	// APIVersion текущая версия API
	APIVersion = "v1"
	
	// APIPrefix префикс для всех API эндпоинтов
	APIPrefix = "/api/v1"
	
	// DefaultTimeout таймаут по умолчанию для HTTP запросов
	DefaultTimeout = 30 * time.Second
	
	// DefaultUserAgent User-Agent по умолчанию
	DefaultUserAgent = "GophKeeper-Client/1.0"
)

// SecretType тип секрета
type SecretType string

const (
	// SecretTypeCredentials логин и пароль
	SecretTypeCredentials SecretType = "credentials"
	
	// SecretTypeText произвольные текстовые данные
	SecretTypeText SecretType = "text"
	
	// SecretTypeBinary бинарные данные
	SecretTypeBinary SecretType = "binary"
	
	// SecretTypeCard данные банковской карты
	SecretTypeCard SecretType = "card"
)

// String возвращает строковое представление типа секрета
func (st SecretType) String() string {
	return string(st)
}

// IsValid проверяет, является ли тип секрета валидным
func (st SecretType) IsValid() bool {
	switch st {
	case SecretTypeCredentials, SecretTypeText, SecretTypeBinary, SecretTypeCard:
		return true
	default:
		return false
	}
}

// OperationType тип операции синхронизации
type OperationType string

const (
	// OperationCreate создание секрета
	OperationCreate OperationType = "create"
	
	// OperationUpdate обновление секрета
	OperationUpdate OperationType = "update"
	
	// OperationDelete удаление секрета
	OperationDelete OperationType = "delete"
)

// String возвращает строковое представление типа операции
func (ot OperationType) String() string {
	return string(ot)
}

// User представляет пользователя системы
type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SecretMetadata базовые метаданные секрета
type SecretMetadata struct {
	ID        uuid.UUID  `json:"id"`
	Type      SecretType `json:"type"`
	Name      string     `json:"name"`
	Metadata  string     `json:"metadata"`
	SyncHash  string     `json:"sync_hash"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// SyncOperation операция синхронизации
type SyncOperation struct {
	ID        uuid.UUID     `json:"id"`
	UserID    uuid.UUID     `json:"user_id"`
	SecretID  uuid.UUID     `json:"secret_id"`
	Operation OperationType `json:"operation"`
	Timestamp time.Time     `json:"timestamp"`
}