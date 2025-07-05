package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ClientConfig конфигурация клиента
type ClientConfig struct {
	ServerURL      string
	ConfigDir      string
	MasterPassword string
}

// SetupClientFlags настраивает флаги для клиента
func SetupClientFlags(cmd *cobra.Command) (*ClientConfig, error) {
	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	defaultConfigDir := filepath.Join(homeDir, ".gophkeeper")

	config := &ClientConfig{}

	cmd.PersistentFlags().StringVar(&config.ServerURL, "server", "http://localhost:8080", "Server URL")
	cmd.PersistentFlags().StringVar(&config.ConfigDir, "config-dir", defaultConfigDir, "Configuration directory")
	cmd.PersistentFlags().StringVar(&config.MasterPassword, "master-password", "", "Master password for encryption")

	return config, nil
}

// GetClientConfig возвращает текущую конфигурацию клиента
func GetClientConfig(serverURL, configDir, masterPassword string) *ClientConfig {
	return &ClientConfig{
		ServerURL:      serverURL,
		ConfigDir:      configDir,
		MasterPassword: masterPassword,
	}
}