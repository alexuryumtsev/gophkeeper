package client

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

func TestNewLocalStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, tempDir, storage.dataDir)

	// Проверяем что директория создана
	assert.DirExists(t, tempDir)
}

func TestNewLocalStorage_CreateDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Удаляем созданную директорию
	os.RemoveAll(tempDir)
	assert.NoDirExists(t, tempDir)

	// NewLocalStorage должен создать директорию
	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)
	assert.DirExists(t, tempDir)
	assert.NotNil(t, storage)
}

func TestLocalStorage_SaveAndGetSecret(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Создаем тестовый секрет
	secretID := uuid.New()
	secret := models.SecretResponse{
		ID:        secretID,
		Name:      "Test Secret",
		Type:      models.SecretTypeCredentials,
		Data:      map[string]any{"username": "user", "password": "pass"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем секрет
	err = storage.SaveSecret(&secret)
	require.NoError(t, err)

	// Получаем секрет
	retrieved := storage.GetSecret(secret.ID)
	require.NotNil(t, retrieved)
	assert.Equal(t, secret.ID, retrieved.ID)
	assert.Equal(t, secret.Name, retrieved.Name)
	assert.Equal(t, secret.Type, retrieved.Type)
	assert.Equal(t, secret.Data, retrieved.Data)
}

func TestLocalStorage_GetSecret_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	nonexistentID := uuid.New()
	secret := storage.GetSecret(nonexistentID)
	assert.Nil(t, secret)
}

func TestLocalStorage_SaveSecrets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	secret1ID := uuid.New()
	secret2ID := uuid.New()
	secrets := []*models.SecretResponse{
		{
			ID:   secret1ID,
			Name: "Secret 1",
			Type: models.SecretTypeCredentials,
			Data: map[string]any{"username": "user1"},
		},
		{
			ID:   secret2ID,
			Name: "Secret 2",
			Type: models.SecretTypeText,
			Data: map[string]any{"text": "some text"},
		},
	}

	err = storage.SaveSecrets(secrets)
	require.NoError(t, err)

	// Проверяем что оба секрета сохранены
	secret1 := storage.GetSecret(secret1ID)
	require.NotNil(t, secret1)
	assert.Equal(t, "Secret 1", secret1.Name)

	secret2 := storage.GetSecret(secret2ID)
	require.NotNil(t, secret2)
	assert.Equal(t, "Secret 2", secret2.Name)
}

func TestLocalStorage_GetSecrets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Сохраняем несколько секретов
	secret1ID := uuid.New()
	secret2ID := uuid.New()
	secret3ID := uuid.New()
	secrets := []*models.SecretResponse{
		{ID: secret1ID, Name: "Secret 1", Type: models.SecretTypeCredentials},
		{ID: secret2ID, Name: "Secret 2", Type: models.SecretTypeText},
		{ID: secret3ID, Name: "Secret 3", Type: models.SecretTypeCard},
	}

	for _, secret := range secrets {
		err = storage.SaveSecret(secret)
		require.NoError(t, err)
	}

	// Получаем все секреты
	allSecrets := storage.GetSecrets()
	assert.Len(t, allSecrets, 3)

	// Проверяем что все секреты присутствуют
	secretIDs := make(map[uuid.UUID]bool)
	for _, secret := range allSecrets {
		secretIDs[secret.ID] = true
	}
	assert.True(t, secretIDs[secret1ID])
	assert.True(t, secretIDs[secret2ID])
	assert.True(t, secretIDs[secret3ID])
}

func TestLocalStorage_DeleteSecret(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Сохраняем секрет
	secretID := uuid.New()
	secret := models.SecretResponse{
		ID:   secretID,
		Name: "Test Secret",
		Type: models.SecretTypeCredentials,
	}
	err = storage.SaveSecret(&secret)
	require.NoError(t, err)

	// Проверяем что секрет существует
	retrieved := storage.GetSecret(secretID)
	require.NotNil(t, retrieved)

	// Удаляем секрет
	err = storage.DeleteSecret(secretID)
	require.NoError(t, err)

	// Проверяем что секрет удален
	deleted := storage.GetSecret(secretID)
	assert.Nil(t, deleted)
}

func TestLocalStorage_DeleteSecrets(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Сохраняем секреты
	secret1ID := uuid.New()
	secret2ID := uuid.New()
	secret3ID := uuid.New()
	secrets := []*models.SecretResponse{
		{ID: secret1ID, Name: "Secret 1"},
		{ID: secret2ID, Name: "Secret 2"},
		{ID: secret3ID, Name: "Secret 3"},
	}

	for _, secret := range secrets {
		err = storage.SaveSecret(secret)
		require.NoError(t, err)
	}

	// Удаляем некоторые секреты
	err = storage.DeleteSecrets([]uuid.UUID{secret1ID, secret3ID})
	require.NoError(t, err)

	// Проверяем что секреты удалены
	deleted1 := storage.GetSecret(secret1ID)
	assert.Nil(t, deleted1)

	deleted3 := storage.GetSecret(secret3ID)
	assert.Nil(t, deleted3)

	// Проверяем что secret-2 остался
	remaining := storage.GetSecret(secret2ID)
	require.NotNil(t, remaining)
}

func TestLocalStorage_GetHashes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Сохраняем секреты
	secret1ID := uuid.New()
	secret2ID := uuid.New()
	secrets := []*models.SecretResponse{
		{ID: secret1ID, Name: "Secret 1", SyncHash: "hash1"},
		{ID: secret2ID, Name: "Secret 2", SyncHash: "hash2"},
	}

	for _, secret := range secrets {
		err = storage.SaveSecret(secret)
		require.NoError(t, err)
	}

	hashes := storage.GetHashes()
	assert.Len(t, hashes, 2)
	assert.Equal(t, "hash1", hashes[secret1ID])
	assert.Equal(t, "hash2", hashes[secret2ID])
}

func TestLocalStorage_SyncTimes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Изначально время синхронизации должно быть нулевым
	lastSync := storage.GetLastSyncTime()
	assert.True(t, lastSync.IsZero())

	// Устанавливаем время синхронизации
	now := time.Now()
	err = storage.SetLastSyncTime(now)
	require.NoError(t, err)

	// Получаем время синхронизации
	retrieved := storage.GetLastSyncTime()

	// Сравниваем с точностью до секунды (из-за сериализации JSON)
	assert.WithinDuration(t, now, retrieved, time.Second)
}
