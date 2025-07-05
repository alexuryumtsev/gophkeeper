package models

import (
	"time"

	"github.com/google/uuid"
)

// SecretType тип секрета
type SecretType string

const (
	SecretTypeCredentials SecretType = "credentials"
	SecretTypeText        SecretType = "text"
	SecretTypeBinary      SecretType = "binary"
	SecretTypeCard        SecretType = "card"
)

// Secret базовая структура для всех типов секретов
type Secret struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Type      SecretType `json:"type" db:"type"`
	Name      string     `json:"name" db:"name"`
	Metadata  string     `json:"metadata" db:"metadata"`
	Data      []byte     `json:"-" db:"data"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	SyncHash  string     `json:"sync_hash" db:"sync_hash"`
}

// Credentials структура для хранения логина и пароля
type Credentials struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	URL      string    `json:"url,omitempty"`
	Metadata string    `json:"metadata,omitempty"`
}

// TextData структура для произвольных текстовых данных
type TextData struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Content  string    `json:"content"`
	Metadata string    `json:"metadata,omitempty"`
}

// BinaryData структура для бинарных данных
type BinaryData struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Data     []byte    `json:"data"`
	Metadata string    `json:"metadata,omitempty"`
}

// CardData структура для данных банковских карт
type CardData struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Number      string    `json:"number"`
	ExpiryMonth int       `json:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year"`
	CVV         string    `json:"cvv"`
	Holder      string    `json:"holder"`
	Bank        string    `json:"bank,omitempty"`
	Metadata    string    `json:"metadata,omitempty"`
}

// SecretRequest запрос на создание/обновление секрета
type SecretRequest struct {
	Type     SecretType `json:"type" binding:"required"`
	Name     string     `json:"name" binding:"required"`
	Data     any        `json:"data" binding:"required"`
	Metadata string     `json:"metadata,omitempty"`
}

// SecretResponse ответ с данными секрета
type SecretResponse struct {
	ID        uuid.UUID  `json:"id"`
	Type      SecretType `json:"type"`
	Name      string     `json:"name"`
	Data      any        `json:"data"`
	Metadata  string     `json:"metadata"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	SyncHash  string     `json:"sync_hash"`
}

// SecretsListResponse ответ со списком секретов
type SecretsListResponse struct {
	Secrets []SecretResponse `json:"secrets"`
	Total   int              `json:"total"`
}
