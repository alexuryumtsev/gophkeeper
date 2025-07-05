package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/uryumtsevaa/gophkeeper/internal/models"
)

// MockConn мок для pgx.Conn
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockConn) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	args := m.Called(ctx, txOptions)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockConn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockConn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	mockArgs := m.Called(ctx, sql, args)
	return mockArgs.Get(0).(pgx.Rows), mockArgs.Error(1)
}

func (m *MockConn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	mockArgs := m.Called(ctx, sql, args)
	return mockArgs.Get(0).(pgx.Row)
}

func (m *MockConn) Close() {
	m.Called()
}

// Вспомогательная функция для создания CommandTag
func NewMockCommandTag(rowsAffected int64) pgconn.CommandTag {
	return pgconn.NewCommandTag(fmt.Sprintf("TEST %d", rowsAffected))
}

// MockRow мок для pgx.Row
type MockRow struct {
	mock.Mock
	scanFunc func(dest ...interface{}) error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	args := m.Called(dest)
	return args.Error(0)
}

// MockRows мок для pgx.Rows
type MockRows struct {
	mock.Mock
	nextFunc  func() bool
	scanFunc  func(dest ...interface{}) error
	closeFunc func()
	data      [][]interface{}
	index     int
}

func (m *MockRows) Next() bool {
	if m.nextFunc != nil {
		return m.nextFunc()
	}
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest...)
	}
	args := m.Called(dest)
	return args.Error(0)
}

func (m *MockRows) Close() {
	if m.closeFunc != nil {
		m.closeFunc()
		return
	}
	m.Called()
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	return NewMockCommandTag(0)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *MockRows) Values() ([]interface{}, error) {
	return nil, nil
}

func (m *MockRows) RawValues() [][]byte {
	return nil
}

func (m *MockRows) Conn() *pgx.Conn {
	return nil
}

// MockPool мок для pgxpool.Pool
type MockPool struct {
	mock.Mock
}

// Ensure MockPool implements DBExecutor
var _ DBExecutor = (*MockPool)(nil)

func (m *MockPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := append([]interface{}{ctx, sql}, arguments...)
	mockArgs := m.Called(args...)
	if mockArgs.Get(0) == nil {
		return NewMockCommandTag(0), mockArgs.Error(1)
	}
	return mockArgs.Get(0).(pgconn.CommandTag), mockArgs.Error(1)
}

func (m *MockPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	mockArgs := append([]interface{}{ctx, sql}, args...)
	result := m.Called(mockArgs...)
	return result.Get(0).(pgx.Rows), result.Error(1)
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	mockArgs := append([]interface{}{ctx, sql}, args...)
	result := m.Called(mockArgs...)
	return result.Get(0).(pgx.Row)
}

func (m *MockPool) Close() {
	m.Called()
}

// MockEncryptor мок для crypto.Encryptor
type MockEncryptor struct {
	mock.Mock
}

func (m *MockEncryptor) Encrypt(data []byte, password string) ([]byte, error) {
	args := m.Called(data, password)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptor) Decrypt(data []byte, password string) ([]byte, error) {
	args := m.Called(data, password)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptor) EncryptString(data string, password string) (string, error) {
	args := m.Called(data, password)
	return args.String(0), args.Error(1)
}

func (m *MockEncryptor) DecryptString(data string, password string) (string, error) {
	args := m.Called(data, password)
	return args.String(0), args.Error(1)
}

func (m *MockEncryptor) GenerateKey(password string, salt []byte) []byte {
	args := m.Called(password, salt)
	return args.Get(0).([]byte)
}

// Тестовые данные
func getTestUserRepo() *models.User {
	return &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func getTestSecret() *models.Secret {
	return &models.Secret{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      models.SecretTypeCredentials,
		Name:      "Test Secret",
		Metadata:  "test metadata",
		Data:      []byte("encrypted_data"),
		SyncHash:  "test_hash",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func getTestSyncOperation() *models.SyncOperation {
	return &models.SyncOperation{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		SecretID:  uuid.New(),
		Operation: models.OperationCreate,
		Timestamp: time.Now(),
	}
}

// Тесты для CreateUser
func TestPostgresRepository_CreateUser(t *testing.T) {
	t.Run("успешное создание пользователя", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testUser := getTestUserRepo()
		mockResult := NewMockCommandTag(1)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			testUser.ID, testUser.Username, testUser.Email, testUser.PasswordHash,
			testUser.CreatedAt, testUser.UpdatedAt).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.CreateUser(ctx, testUser)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка базы данных", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testUser := getTestUserRepo()
		dbError := fmt.Errorf("database error")

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(NewMockCommandTag(0), dbError)

		ctx := context.Background()
		err := repo.CreateUser(ctx, testUser)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для GetUserByUsername
func TestPostgresRepository_GetUserByUsername(t *testing.T) {
	t.Run("успешное получение пользователя", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testUser := getTestUserRepo()
		mockRow := &MockRow{}

		mockRow.scanFunc = func(dest ...interface{}) error {
			// Заполняем значениями тестового пользователя
			if len(dest) >= 6 {
				*(dest[0].(*uuid.UUID)) = testUser.ID
				*(dest[1].(*string)) = testUser.Username
				*(dest[2].(*string)) = testUser.Email
				*(dest[3].(*string)) = testUser.PasswordHash
				*(dest[4].(*time.Time)) = testUser.CreatedAt
				*(dest[5].(*time.Time)) = testUser.UpdatedAt
			}
			return nil
		}

		mockPool.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), testUser.Username).Return(mockRow)

		ctx := context.Background()
		result, err := repo.GetUserByUsername(ctx, testUser.Username)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.Username, result.Username)
		assert.Equal(t, testUser.Email, result.Email)
		mockPool.AssertExpectations(t)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		mockRow := &MockRow{}
		mockRow.scanFunc = func(dest ...interface{}) error {
			return pgx.ErrNoRows
		}

		mockPool.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), "nonexistent").Return(mockRow)

		ctx := context.Background()
		result, err := repo.GetUserByUsername(ctx, "nonexistent")

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка базы данных", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		mockRow := &MockRow{}
		dbError := fmt.Errorf("database error")
		mockRow.scanFunc = func(dest ...interface{}) error {
			return dbError
		}

		mockPool.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), "testuser").Return(mockRow)

		ctx := context.Background()
		result, err := repo.GetUserByUsername(ctx, "testuser")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get user by username")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для GetUserByID
func TestPostgresRepository_GetUserByID(t *testing.T) {
	t.Run("успешное получение пользователя по ID", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testUser := getTestUserRepo()
		mockRow := &MockRow{}

		mockRow.scanFunc = func(dest ...interface{}) error {
			if len(dest) >= 6 {
				*(dest[0].(*uuid.UUID)) = testUser.ID
				*(dest[1].(*string)) = testUser.Username
				*(dest[2].(*string)) = testUser.Email
				*(dest[3].(*string)) = testUser.PasswordHash
				*(dest[4].(*time.Time)) = testUser.CreatedAt
				*(dest[5].(*time.Time)) = testUser.UpdatedAt
			}
			return nil
		}

		mockPool.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), testUser.ID).Return(mockRow)

		ctx := context.Background()
		result, err := repo.GetUserByID(ctx, testUser.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.ID)
		assert.Equal(t, testUser.Username, result.Username)
		mockPool.AssertExpectations(t)
	})
}

// Тесты для CreateSecret
func TestPostgresRepository_CreateSecret(t *testing.T) {
	t.Run("успешное создание секрета", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSecret := getTestSecret()
		mockResult := NewMockCommandTag(1)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			testSecret.ID, testSecret.UserID, testSecret.Type, testSecret.Name,
			testSecret.Metadata, testSecret.Data, testSecret.SyncHash,
			testSecret.CreatedAt, testSecret.UpdatedAt).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.CreateSecret(ctx, testSecret)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка базы данных при создании секрета", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSecret := getTestSecret()
		dbError := fmt.Errorf("database error")

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(NewMockCommandTag(0), dbError)

		ctx := context.Background()
		err := repo.CreateSecret(ctx, testSecret)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create secret")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для GetSecretsByUserID
func TestPostgresRepository_GetSecretsByUserID(t *testing.T) {
	t.Run("успешное получение секретов пользователя", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSecret1 := getTestSecret()
		testSecret2 := getTestSecret()
		userID := uuid.New()

		mockRows := &MockRows{}
		callIndex := 0
		mockRows.nextFunc = func() bool {
			callIndex++
			return callIndex <= 2 // Возвращаем true для двух итераций
		}

		mockRows.scanFunc = func(dest ...interface{}) error {
			var secret *models.Secret
			if callIndex == 1 {
				secret = testSecret1
			} else {
				secret = testSecret2
			}

			if len(dest) >= 9 {
				*(dest[0].(*uuid.UUID)) = secret.ID
				*(dest[1].(*uuid.UUID)) = secret.UserID
				*(dest[2].(*models.SecretType)) = secret.Type
				*(dest[3].(*string)) = secret.Name
				*(dest[4].(*string)) = secret.Metadata
				*(dest[5].(*[]byte)) = secret.Data
				*(dest[6].(*string)) = secret.SyncHash
				*(dest[7].(*time.Time)) = secret.CreatedAt
				*(dest[8].(*time.Time)) = secret.UpdatedAt
			}
			return nil
		}

		mockRows.closeFunc = func() {}

		mockPool.On("Query", mock.Anything, mock.AnythingOfType("string"), userID).Return(mockRows, nil)

		ctx := context.Background()
		result, err := repo.GetSecretsByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка при выполнении запроса", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		userID := uuid.New()
		dbError := fmt.Errorf("query error")

		mockPool.On("Query", mock.Anything, mock.AnythingOfType("string"), userID).Return((*MockRows)(nil), dbError)

		ctx := context.Background()
		result, err := repo.GetSecretsByUserID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get secrets")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для GetSecretByID
func TestPostgresRepository_GetSecretByID(t *testing.T) {
	t.Run("успешное получение секрета по ID", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSecret := getTestSecret()
		mockRow := &MockRow{}

		mockRow.scanFunc = func(dest ...interface{}) error {
			if len(dest) >= 9 {
				*(dest[0].(*uuid.UUID)) = testSecret.ID
				*(dest[1].(*uuid.UUID)) = testSecret.UserID
				*(dest[2].(*models.SecretType)) = testSecret.Type
				*(dest[3].(*string)) = testSecret.Name
				*(dest[4].(*string)) = testSecret.Metadata
				*(dest[5].(*[]byte)) = testSecret.Data
				*(dest[6].(*string)) = testSecret.SyncHash
				*(dest[7].(*time.Time)) = testSecret.CreatedAt
				*(dest[8].(*time.Time)) = testSecret.UpdatedAt
			}
			return nil
		}

		mockPool.On("QueryRow", mock.Anything, mock.AnythingOfType("string"),
			testSecret.ID, testSecret.UserID).Return(mockRow)

		ctx := context.Background()
		result, err := repo.GetSecretByID(ctx, testSecret.ID, testSecret.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testSecret.ID, result.ID)
		assert.Equal(t, testSecret.Name, result.Name)
		mockPool.AssertExpectations(t)
	})

	t.Run("секрет не найден", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		mockRow := &MockRow{}
		mockRow.scanFunc = func(dest ...interface{}) error {
			return pgx.ErrNoRows
		}

		secretID := uuid.New()
		userID := uuid.New()
		mockPool.On("QueryRow", mock.Anything, mock.AnythingOfType("string"),
			secretID, userID).Return(mockRow)

		ctx := context.Background()
		result, err := repo.GetSecretByID(ctx, secretID, userID)

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockPool.AssertExpectations(t)
	})
}

// Тесты для UpdateSecret
func TestPostgresRepository_UpdateSecret(t *testing.T) {
	t.Run("успешное обновление секрета", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSecret := getTestSecret()
		mockResult := NewMockCommandTag(1)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			testSecret.ID, testSecret.UserID, testSecret.Name, testSecret.Metadata,
			testSecret.Data, testSecret.SyncHash, testSecret.UpdatedAt).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.UpdateSecret(ctx, testSecret)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("секрет не найден при обновлении", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSecret := getTestSecret()
		mockResult := NewMockCommandTag(0)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.UpdateSecret(ctx, testSecret)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found or access denied")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для DeleteSecret
func TestPostgresRepository_DeleteSecret(t *testing.T) {
	t.Run("успешное удаление секрета", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		secretID := uuid.New()
		userID := uuid.New()
		mockResult := NewMockCommandTag(1)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			secretID, userID).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.DeleteSecret(ctx, secretID, userID)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("секрет не найден при удалении", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		secretID := uuid.New()
		userID := uuid.New()
		mockResult := NewMockCommandTag(0)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			secretID, userID).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.DeleteSecret(ctx, secretID, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found or access denied")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для CreateSyncOperation
func TestPostgresRepository_CreateSyncOperation(t *testing.T) {
	t.Run("успешное создание операции синхронизации", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSyncOp := getTestSyncOperation()
		mockResult := NewMockCommandTag(1)

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			testSyncOp.ID, testSyncOp.UserID, testSyncOp.SecretID,
			testSyncOp.Operation, testSyncOp.Timestamp).Return(mockResult, nil)

		ctx := context.Background()
		err := repo.CreateSyncOperation(ctx, testSyncOp)

		assert.NoError(t, err)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка базы данных при создании операции синхронизации", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSyncOp := getTestSyncOperation()
		dbError := fmt.Errorf("database error")

		mockPool.On("Exec", mock.Anything, mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(NewMockCommandTag(0), dbError)

		ctx := context.Background()
		err := repo.CreateSyncOperation(ctx, testSyncOp)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create sync operation")
		mockPool.AssertExpectations(t)
	})
}

// Тесты для GetSyncOperationsAfter
func TestPostgresRepository_GetSyncOperationsAfter(t *testing.T) {
	t.Run("успешное получение операций синхронизации", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		testSyncOp := getTestSyncOperation()
		userID := uuid.New()
		after := time.Now().Add(-time.Hour)

		mockRows := &MockRows{}
		callIndex := 0
		mockRows.nextFunc = func() bool {
			callIndex++
			return callIndex <= 1 // Возвращаем true для одной итерации
		}

		mockRows.scanFunc = func(dest ...interface{}) error {
			if len(dest) >= 5 {
				*(dest[0].(*uuid.UUID)) = testSyncOp.ID
				*(dest[1].(*uuid.UUID)) = testSyncOp.UserID
				*(dest[2].(*sql.NullString)) = sql.NullString{String: testSyncOp.SecretID.String(), Valid: true}
				*(dest[3].(*models.OperationType)) = testSyncOp.Operation
				*(dest[4].(*time.Time)) = testSyncOp.Timestamp
			}
			return nil
		}

		mockRows.closeFunc = func() {}

		mockPool.On("Query", mock.Anything, mock.AnythingOfType("string"), userID, after).Return(mockRows, nil)

		ctx := context.Background()
		result, err := repo.GetSyncOperationsAfter(ctx, userID, after)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, testSyncOp.Operation, result[0].Operation)
		mockPool.AssertExpectations(t)
	})

	t.Run("ошибка при получении операций синхронизации", func(t *testing.T) {
		mockPool := new(MockPool)
		mockEncryptor := new(MockEncryptor)
		repo := NewPostgresRepository(mockPool, mockEncryptor)

		userID := uuid.New()
		after := time.Now().Add(-time.Hour)
		dbError := fmt.Errorf("query error")

		mockPool.On("Query", mock.Anything, mock.AnythingOfType("string"), userID, after).Return((*MockRows)(nil), dbError)

		ctx := context.Background()
		result, err := repo.GetSyncOperationsAfter(ctx, userID, after)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get sync operations")
		mockPool.AssertExpectations(t)
	})
}
