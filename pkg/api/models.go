package api

import (
	"time"

	"github.com/google/uuid"
)

// Authentication models

// RegisterRequest запрос на регистрацию нового пользователя
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest запрос на авторизацию
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse ответ на успешную авторизацию/регистрацию
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Secret data types

// Credentials данные для логина и пароля
type Credentials struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	URL      string    `json:"url,omitempty"`
	Metadata string    `json:"metadata,omitempty"`
}

// TextData произвольные текстовые данные
type TextData struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Content  string    `json:"content"`
	Metadata string    `json:"metadata,omitempty"`
}

// BinaryData бинарные данные
type BinaryData struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Data     []byte    `json:"data"`
	Metadata string    `json:"metadata,omitempty"`
}

// CardData данные банковской карты
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

// Secret models

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
	SyncHash  string     `json:"sync_hash"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// SecretsListResponse ответ со списком секретов
type SecretsListResponse struct {
	Secrets []SecretResponse `json:"secrets"`
	Total   int              `json:"total"`
}

// Sync models

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

// Utility methods

// Validate проверяет валидность запроса на регистрацию
func (r *RegisterRequest) Validate() error {
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return &ValidationError{Field: "username", Message: "Username must be between 3 and 50 characters"}
	}
	if len(r.Password) < 8 {
		return &ValidationError{Field: "password", Message: "Password must be at least 8 characters"}
	}
	return nil
}

// Validate проверяет валидность запроса на создание секрета
func (r *SecretRequest) Validate() error {
	if !r.Type.IsValid() {
		return &ValidationError{Field: "type", Message: "Invalid secret type"}
	}
	if r.Name == "" {
		return &ValidationError{Field: "name", Message: "Name is required"}
	}
	if r.Data == nil {
		return &ValidationError{Field: "data", Message: "Data is required"}
	}
	return nil
}

// GetCredentials возвращает данные как Credentials (если возможно)
func (r *SecretResponse) GetCredentials() (*Credentials, bool) {
	if r.Type != SecretTypeCredentials {
		return nil, false
	}
	if creds, ok := r.Data.(map[string]any); ok {
		credentials := &Credentials{
			ID:   r.ID,
			Name: r.Name,
		}
		if username, ok := creds["username"].(string); ok {
			credentials.Username = username
		}
		if password, ok := creds["password"].(string); ok {
			credentials.Password = password
		}
		if url, ok := creds["url"].(string); ok {
			credentials.URL = url
		}
		credentials.Metadata = r.Metadata
		return credentials, true
	}
	return nil, false
}

// GetTextData возвращает данные как TextData (если возможно)
func (r *SecretResponse) GetTextData() (*TextData, bool) {
	if r.Type != SecretTypeText {
		return nil, false
	}
	if textData, ok := r.Data.(map[string]any); ok {
		text := &TextData{
			ID:   r.ID,
			Name: r.Name,
		}
		if content, ok := textData["content"].(string); ok {
			text.Content = content
		}
		text.Metadata = r.Metadata
		return text, true
	}
	return nil, false
}

// GetCardData возвращает данные как CardData (если возможно)
func (r *SecretResponse) GetCardData() (*CardData, bool) {
	if r.Type != SecretTypeCard {
		return nil, false
	}
	if cardData, ok := r.Data.(map[string]any); ok {
		card := &CardData{
			ID:   r.ID,
			Name: r.Name,
		}
		if number, ok := cardData["number"].(string); ok {
			card.Number = number
		}
		if holder, ok := cardData["holder"].(string); ok {
			card.Holder = holder
		}
		if cvv, ok := cardData["cvv"].(string); ok {
			card.CVV = cvv
		}
		if bank, ok := cardData["bank"].(string); ok {
			card.Bank = bank
		}
		if expiryMonth, ok := cardData["expiry_month"].(float64); ok {
			card.ExpiryMonth = int(expiryMonth)
		}
		if expiryYear, ok := cardData["expiry_year"].(float64); ok {
			card.ExpiryYear = int(expiryYear)
		}
		card.Metadata = r.Metadata
		return card, true
	}
	return nil, false
}