package storage

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database представляет подключение к базе данных
type Database struct {
	Pool *pgxpool.Pool
}

// Config конфигурация базы данных
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Validate проверяет корректность конфигурации
func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}
	if c.User == "" {
		return fmt.Errorf("user is required")
	}
	if c.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	for _, mode := range validSSLModes {
		if c.SSLMode == mode {
			return nil
		}
	}
	return fmt.Errorf("invalid SSL mode: %s", c.SSLMode)
}

// DSN возвращает строку подключения к базе данных
func (c Config) DSN() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.DBName, c.SSLMode)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// MigrateURL возвращает URL для golang-migrate
func (c Config) MigrateURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}

// String возвращает строковое представление конфигурации (с замаскированным паролем)
func (c Config) String() string {
	password := c.Password
	if password != "" {
		password = "***"
	}
	return fmt.Sprintf("Config{Host: %s, Port: %d, User: %s, Password: %s, DBName: %s, SSLMode: %s}",
		c.Host, c.Port, c.User, password, c.DBName, c.SSLMode)
}

// NewDatabase создает новое подключение к базе данных
func NewDatabase(ctx context.Context, config Config) (*Database, error) {
	dsn := config.DSN()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

// Close закрывает подключение к базе данных
func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// RunMigrations выполняет миграции базы данных
func (db *Database) RunMigrations(migrationsPath string, config Config) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		config.MigrateURL(),
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
