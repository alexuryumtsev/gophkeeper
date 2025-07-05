package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: Config{
				Host:     "",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
		{
			name: "zero port",
			config: Config{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Password: "password",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
		{
			name: "empty user",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "",
				Password: "password",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
		{
			name: "empty database name",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
		{
			name: "invalid SSL mode",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "test",
				SSLMode:  "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid SSL mode - require",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "test",
				SSLMode:  "require",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "basic config",
			config: Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "test",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=test sslmode=disable",
		},
		{
			name: "config without password",
			config: Config{
				Host:    "localhost",
				Port:    5432,
				User:    "postgres",
				DBName:  "test",
				SSLMode: "disable",
			},
			expected: "host=localhost port=5432 user=postgres dbname=test sslmode=disable",
		},
		{
			name: "config with special characters",
			config: Config{
				Host:     "test-host",
				Port:     3306,
				User:     "test_user",
				Password: "pass@word!",
				DBName:   "test-db",
				SSLMode:  "require",
			},
			expected: "host=test-host port=3306 user=test_user password=pass@word! dbname=test-db sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

// Интеграционные тесты требуют запущенной БД PostgreSQL
// Они выполняются только если установлена переменная окружения TEST_DATABASE_URL
func TestNewDatabase_Integration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "test",
		SSLMode:  "disable",
	}

	// Валидируем конфигурацию
	err := config.Validate()
	require.NoError(t, err)

	ctx := context.Background()
	db, err := NewDatabase(ctx, config)
	
	if err != nil {
		// Если не удается подключиться к БД, пропускаем тест
		t.Skipf("Cannot connect to database: %v", err)
	}

	require.NotNil(t, db)
	require.NotNil(t, db.Pool)

	// Проверяем что можем выполнить простой запрос
	var result int
	err = db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	// Закрываем соединение
	db.Close()
}

func TestDatabase_Close(t *testing.T) {
	// Создаем БД с невалидной конфигурацией для теста
	db := &Database{
		Pool: nil,
	}

	// Close не должен паниковать даже с nil pool
	assert.NotPanics(t, func() {
		db.Close()
	})
}

func TestRunMigrations_NoMigrationsDir(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "test",
		SSLMode:  "disable",
	}

	db := &Database{}

	// Тест с несуществующей директорией миграций
	err := db.RunMigrations("/nonexistent/path", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestConfig_String(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DBName:   "test",
		SSLMode:  "disable",
	}

	str := config.String()
	
	// Проверяем что строка содержит основную информацию
	assert.Contains(t, str, "localhost")
	assert.Contains(t, str, "5432")
	assert.Contains(t, str, "postgres")
	assert.Contains(t, str, "test")
	assert.Contains(t, str, "disable")
	
	// Проверяем что пароль замаскирован
	assert.NotContains(t, str, "secret")
	assert.Contains(t, str, "***")
}