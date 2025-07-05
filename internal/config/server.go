package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uryumtsevaa/gophkeeper/internal/server"
	"github.com/uryumtsevaa/gophkeeper/internal/storage"
)

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port      int
	JWTSecret string
	Database  storage.Config
}

// SetupServerFlags настраивает флаги для сервера
func SetupServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().IntP("port", "p", 8080, "Server port")
	cmd.PersistentFlags().String("jwt-secret", "", "JWT secret key")
	cmd.PersistentFlags().String("db-host", "localhost", "Database host")
	cmd.PersistentFlags().IntP("db-port", "", 5432, "Database port")
	cmd.PersistentFlags().String("db-user", "postgres", "Database user")
	cmd.PersistentFlags().String("db-password", "", "Database password")
	cmd.PersistentFlags().String("db-name", "gophkeeper", "Database name")
	cmd.PersistentFlags().String("db-sslmode", "disable", "Database SSL mode")

	viper.BindPFlag("port", cmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("jwt_secret", cmd.PersistentFlags().Lookup("jwt-secret"))
	viper.BindPFlag("database.host", cmd.PersistentFlags().Lookup("db-host"))
	viper.BindPFlag("database.port", cmd.PersistentFlags().Lookup("db-port"))
	viper.BindPFlag("database.user", cmd.PersistentFlags().Lookup("db-user"))
	viper.BindPFlag("database.password", cmd.PersistentFlags().Lookup("db-password"))
	viper.BindPFlag("database.name", cmd.PersistentFlags().Lookup("db-name"))
	viper.BindPFlag("database.sslmode", cmd.PersistentFlags().Lookup("db-sslmode"))
}

// InitServerConfig инициализирует конфигурацию сервера
func InitServerConfig(configFile string) {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Читаем переменные окружения
	viper.SetEnvPrefix("GOPHKEEPER")
	viper.AutomaticEnv()

	// Устанавливаем значения по умолчанию
	viper.SetDefault("port", 8080)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "gophkeeper")
	viper.SetDefault("database.sslmode", "disable")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// LoadServerConfig загружает конфигурацию сервера
func LoadServerConfig() server.Config {
	// JWT secret обязателен
	jwtSecret := viper.GetString("jwt_secret")
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Fatal("JWT secret is required. Set it via --jwt-secret flag or JWT_SECRET environment variable")
		}
	}

	dbPort := viper.GetInt("database.port")
	if dbPort == 0 {
		if portStr := os.Getenv("DB_PORT"); portStr != "" {
			if port, err := strconv.Atoi(portStr); err == nil {
				dbPort = port
			}
		}
		if dbPort == 0 {
			dbPort = 5432
		}
	}

	return server.Config{
		Port:      viper.GetInt("port"),
		JWTSecret: jwtSecret,
		Database: storage.Config{
			Host:     getStringWithEnv("database.host", "DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getStringWithEnv("database.user", "DB_USER", "postgres"),
			Password: getStringWithEnv("database.password", "DB_PASSWORD", ""),
			DBName:   getStringWithEnv("database.name", "DB_NAME", "gophkeeper"),
			SSLMode:  getStringWithEnv("database.sslmode", "DB_SSLMODE", "disable"),
		},
	}
}

// getStringWithEnv получает значение из viper или переменной окружения
func getStringWithEnv(viperKey, envKey, defaultValue string) string {
	value := viper.GetString(viperKey)
	if value == "" {
		value = os.Getenv(envKey)
	}
	if value == "" {
		value = defaultValue
	}
	return value
}