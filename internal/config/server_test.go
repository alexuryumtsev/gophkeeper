package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/uryumtsevaa/gophkeeper/internal/server"
)

func TestSetupServerFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	SetupServerFlags(cmd)

	// Проверяем что флаги добавлены
	assert.NotNil(t, cmd.PersistentFlags().Lookup("port"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("jwt-secret"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("db-host"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("db-port"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("db-user"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("db-password"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("db-name"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("db-sslmode"))
}

func TestInitServerConfig(t *testing.T) {
	// Очищаем viper перед тестом
	viper.Reset()

	// Тест с пустым файлом конфигурации
	InitServerConfig("")

	// Проверяем значения по умолчанию
	assert.Equal(t, 8080, viper.GetInt("port"))
	assert.Equal(t, "localhost", viper.GetString("database.host"))
	assert.Equal(t, 5432, viper.GetInt("database.port"))
	assert.Equal(t, "postgres", viper.GetString("database.user"))
	assert.Equal(t, "gophkeeper", viper.GetString("database.name"))
	assert.Equal(t, "disable", viper.GetString("database.sslmode"))
}

func TestLoadServerConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		viperValues map[string]any
		expectError bool
		expected    func() server.Config
	}{
		{
			name: "valid config with JWT secret from env",
			envVars: map[string]string{
				"JWT_SECRET": "test-secret",
			},
			viperValues: map[string]any{
				"port": 9000,
			},
			expected: func() server.Config {
				return server.Config{
					Port:      9000,
					JWTSecret: "test-secret",
					Database: struct {
						Host     string
						Port     int
						User     string
						Password string
						DBName   string
						SSLMode  string
					}{
						Host:     "localhost",
						Port:     5432,
						User:     "postgres",
						Password: "",
						DBName:   "gophkeeper",
						SSLMode:  "disable",
					},
				}
			},
		},
		{
			name: "valid config with JWT secret from viper",
			viperValues: map[string]any{
				"jwt_secret":      "viper-secret",
				"port":            8080,
				"database.host":   "test-host",
				"database.port":   3306,
				"database.user":   "test-user",
				"database.name":   "test-db",
				"database.sslmode": "require",
			},
			expected: func() server.Config {
				return server.Config{
					Port:      8080,
					JWTSecret: "viper-secret",
					Database: struct {
						Host     string
						Port     int
						User     string
						Password string
						DBName   string
						SSLMode  string
					}{
						Host:     "test-host",
						Port:     3306,
						User:     "test-user",
						Password: "",
						DBName:   "test-db",
						SSLMode:  "require",
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем viper перед каждым тестом
			viper.Reset()

			// Устанавливаем переменные окружения
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Устанавливаем значения в viper
			for key, value := range tt.viperValues {
				viper.Set(key, value)
			}

			// Устанавливаем значения по умолчанию как в InitServerConfig
			viper.SetDefault("port", 8080)
			viper.SetDefault("database.host", "localhost")
			viper.SetDefault("database.port", 5432)
			viper.SetDefault("database.user", "postgres")
			viper.SetDefault("database.name", "gophkeeper")
			viper.SetDefault("database.sslmode", "disable")

			if tt.expectError {
				assert.Panics(t, func() {
					LoadServerConfig()
				})
			} else {
				config := LoadServerConfig()
				expected := tt.expected()
				assert.Equal(t, expected, config)
			}

			// Очищаем переменные окружения
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestLoadServerConfigNoJWTSecret(t *testing.T) {
	// Очищаем viper и переменные окружения
	viper.Reset()
	os.Unsetenv("JWT_SECRET")

	// Устанавливаем значения по умолчанию
	viper.SetDefault("port", 8080)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "gophkeeper")
	viper.SetDefault("database.sslmode", "disable")

	// Должен завершиться с ошибкой, если JWT secret не установлен
	// Используем defer для восстановления после возможного log.Fatal
	defer func() {
		if r := recover(); r != nil {
			// Ожидаем panic от log.Fatal
			assert.Contains(t, fmt.Sprintf("%v", r), "JWT secret is required")
		}
	}()

	// Этот вызов должен вызвать log.Fatal
	LoadServerConfig()
	t.Error("Expected LoadServerConfig to call log.Fatal, but it didn't")
}

func TestGetStringWithEnv(t *testing.T) {
	tests := []struct {
		name         string
		viperKey     string
		envKey       string
		defaultValue string
		viperValue   string
		envValue     string
		expected     string
	}{
		{
			name:         "viper value exists",
			viperKey:     "test.key",
			envKey:       "TEST_ENV",
			defaultValue: "default",
			viperValue:   "viper-value",
			envValue:     "",
			expected:     "viper-value",
		},
		{
			name:         "env value exists when viper is empty",
			viperKey:     "test.key",
			envKey:       "TEST_ENV",
			defaultValue: "default",
			viperValue:   "",
			envValue:     "env-value",
			expected:     "env-value",
		},
		{
			name:         "default value when both are empty",
			viperKey:     "test.key",
			envKey:       "TEST_ENV",
			defaultValue: "default",
			viperValue:   "",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем viper перед каждым тестом
			viper.Reset()

			// Устанавливаем значения
			if tt.viperValue != "" {
				viper.Set(tt.viperKey, tt.viperValue)
			}
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
			}

			result := getStringWithEnv(tt.viperKey, tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)

			// Очищаем переменную окружения
			if tt.envValue != "" {
				os.Unsetenv(tt.envKey)
			}
		})
	}
}