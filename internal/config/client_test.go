package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupClientFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	config, err := SetupClientFlags(cmd)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Проверяем что флаги добавлены
	assert.NotNil(t, cmd.PersistentFlags().Lookup("server"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config-dir"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("master-password"))

	// Проверяем значения по умолчанию
	assert.Equal(t, "http://localhost:8080", config.ServerURL)
	assert.Equal(t, "", config.MasterPassword)

	// Проверяем что config-dir содержит .gophkeeper
	assert.Contains(t, config.ConfigDir, ".gophkeeper")
}

func TestSetupClientFlagsWithHomeDir(t *testing.T) {
	// Сохраняем оригинальную переменную окружения
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Устанавливаем тестовую домашнюю директорию
	testHome := "/tmp/test-home"
	os.Setenv("HOME", testHome)

	cmd := &cobra.Command{
		Use: "test",
	}

	config, err := SetupClientFlags(cmd)
	require.NoError(t, err)

	expectedConfigDir := filepath.Join(testHome, ".gophkeeper")
	assert.Equal(t, expectedConfigDir, config.ConfigDir)
}

func TestGetClientConfig(t *testing.T) {
	tests := []struct {
		name           string
		serverURL      string
		configDir      string
		masterPassword string
		expected       *ClientConfig
	}{
		{
			name:           "all parameters provided",
			serverURL:      "https://api.example.com",
			configDir:      "/custom/config",
			masterPassword: "secret123",
			expected: &ClientConfig{
				ServerURL:      "https://api.example.com",
				ConfigDir:      "/custom/config",
				MasterPassword: "secret123",
			},
		},
		{
			name:           "empty parameters",
			serverURL:      "",
			configDir:      "",
			masterPassword: "",
			expected: &ClientConfig{
				ServerURL:      "",
				ConfigDir:      "",
				MasterPassword: "",
			},
		},
		{
			name:           "partial parameters",
			serverURL:      "http://localhost:9000",
			configDir:      "/home/user/.config",
			masterPassword: "",
			expected: &ClientConfig{
				ServerURL:      "http://localhost:9000",
				ConfigDir:      "/home/user/.config",
				MasterPassword: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetClientConfig(tt.serverURL, tt.configDir, tt.masterPassword)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClientConfigStruct(t *testing.T) {
	config := &ClientConfig{
		ServerURL:      "https://example.com",
		ConfigDir:      "/test/config",
		MasterPassword: "password",
	}

	// Тест что структура правильно инициализируется
	assert.Equal(t, "https://example.com", config.ServerURL)
	assert.Equal(t, "/test/config", config.ConfigDir)
	assert.Equal(t, "password", config.MasterPassword)

	// Тест что можно изменять значения
	config.ServerURL = "https://new-url.com"
	config.MasterPassword = "new-password"

	assert.Equal(t, "https://new-url.com", config.ServerURL)
	assert.Equal(t, "new-password", config.MasterPassword)
	assert.Equal(t, "/test/config", config.ConfigDir) // не изменился
}