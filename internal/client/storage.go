package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// LocalStorage локальное хранилище для клиента
type LocalStorage struct {
	mu       sync.RWMutex
	dataDir  string
	secrets  map[uuid.UUID]*models.SecretResponse
	lastSync time.Time
}

// NewLocalStorage создает новое локальное хранилище
func NewLocalStorage(dataDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	storage := &LocalStorage{
		dataDir: dataDir,
		secrets: make(map[uuid.UUID]*models.SecretResponse),
	}

	// Загружаем данные из файла
	if err := storage.load(); err != nil {
		// Если файл не существует, это нормально для первого запуска
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load data: %w", err)
		}
	}

	return storage, nil
}

// GetSecrets возвращает все секреты
func (s *LocalStorage) GetSecrets() []*models.SecretResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secrets := make([]*models.SecretResponse, 0, len(s.secrets))
	for _, secret := range s.secrets {
		secrets = append(secrets, secret)
	}

	return secrets
}

// GetSecret возвращает секрет по ID
func (s *LocalStorage) GetSecret(id uuid.UUID) *models.SecretResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.secrets[id]
}

// SaveSecret сохраняет секрет
func (s *LocalStorage) SaveSecret(secret *models.SecretResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.secrets[secret.ID] = secret
	return s.save()
}

// SaveSecrets сохраняет множество секретов
func (s *LocalStorage) SaveSecrets(secrets []*models.SecretResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, secret := range secrets {
		s.secrets[secret.ID] = secret
	}

	return s.save()
}

// DeleteSecret удаляет секрет
func (s *LocalStorage) DeleteSecret(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.secrets, id)
	return s.save()
}

// DeleteSecrets удаляет множество секретов
func (s *LocalStorage) DeleteSecrets(ids []uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		delete(s.secrets, id)
	}

	return s.save()
}

// GetHashes возвращает хеши всех секретов для синхронизации
func (s *LocalStorage) GetHashes() map[uuid.UUID]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hashes := make(map[uuid.UUID]string)
	for id, secret := range s.secrets {
		hashes[id] = secret.SyncHash
	}

	return hashes
}

// GetLastSyncTime возвращает время последней синхронизации
func (s *LocalStorage) GetLastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastSync
}

// SetLastSyncTime устанавливает время последней синхронизации
func (s *LocalStorage) SetLastSyncTime(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastSync = t
	return s.save()
}

// storageData структура для сохранения в файл
type storageData struct {
	Secrets  map[uuid.UUID]*models.SecretResponse `json:"secrets"`
	LastSync time.Time                            `json:"last_sync"`
}

// save сохраняет данные в файл
func (s *LocalStorage) save() error {
	data := storageData{
		Secrets:  s.secrets,
		LastSync: s.lastSync,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	filePath := filepath.Join(s.dataDir, "secrets.json")
	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// load загружает данные из файла
func (s *LocalStorage) load() error {
	filePath := filepath.Join(s.dataDir, "secrets.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var storageData storageData
	if err := json.Unmarshal(data, &storageData); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	s.secrets = storageData.Secrets
	if s.secrets == nil {
		s.secrets = make(map[uuid.UUID]*models.SecretResponse)
	}
	s.lastSync = storageData.LastSync

	return nil
}
