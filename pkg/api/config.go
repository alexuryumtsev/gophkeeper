package api

import (
	"fmt"
	"net/url"
	"time"
)

// Config конфигурация клиента GophKeeper
type Config struct {
	// BaseURL базовый URL сервера GophKeeper
	BaseURL string
	
	// Timeout таймаут для HTTP запросов
	Timeout time.Duration
	
	// UserAgent User-Agent заголовок для запросов
	UserAgent string
	
	// MaxRetries максимальное количество повторных попыток
	MaxRetries int
	
	// RetryDelay задержка между повторными попытками
	RetryDelay time.Duration
	
	// InsecureSkipVerify пропускать проверку TLS сертификатов (только для разработки)
	InsecureSkipVerify bool
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		BaseURL:            "http://localhost:8080",
		Timeout:            DefaultTimeout,
		UserAgent:          DefaultUserAgent,
		MaxRetries:         3,
		RetryDelay:         time.Second,
		InsecureSkipVerify: false,
	}
}

// NewConfig создает новую конфигурацию с заданным базовым URL
func NewConfig(baseURL string) *Config {
	config := DefaultConfig()
	config.BaseURL = baseURL
	return config
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	
	// Проверяем, что URL можно распарсить
	parsedURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}
	
	if parsedURL.Scheme == "" {
		return fmt.Errorf("base URL must include scheme (http:// or https://)")
	}
	
	if parsedURL.Host == "" {
		return fmt.Errorf("base URL must include host")
	}
	
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if c.UserAgent == "" {
		return fmt.Errorf("user agent is required")
	}
	
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	
	if c.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}
	
	return nil
}

// WithTimeout устанавливает таймаут и возвращает конфигурацию
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithUserAgent устанавливает User-Agent и возвращает конфигурацию
func (c *Config) WithUserAgent(userAgent string) *Config {
	c.UserAgent = userAgent
	return c
}

// WithRetry устанавливает параметры повторных попыток и возвращает конфигурацию
func (c *Config) WithRetry(maxRetries int, retryDelay time.Duration) *Config {
	c.MaxRetries = maxRetries
	c.RetryDelay = retryDelay
	return c
}

// WithInsecureSkipVerify включает/отключает проверку TLS сертификатов и возвращает конфигурацию
func (c *Config) WithInsecureSkipVerify(skip bool) *Config {
	c.InsecureSkipVerify = skip
	return c
}

// Copy создает копию конфигурации
func (c *Config) Copy() *Config {
	return &Config{
		BaseURL:            c.BaseURL,
		Timeout:            c.Timeout,
		UserAgent:          c.UserAgent,
		MaxRetries:         c.MaxRetries,
		RetryDelay:         c.RetryDelay,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}
}

// String возвращает строковое представление конфигурации
func (c *Config) String() string {
	return fmt.Sprintf("Config{BaseURL: %s, Timeout: %v, UserAgent: %s, MaxRetries: %d, RetryDelay: %v, InsecureSkipVerify: %t}",
		c.BaseURL, c.Timeout, c.UserAgent, c.MaxRetries, c.RetryDelay, c.InsecureSkipVerify)
}

// Preset configurations

// LocalConfig возвращает конфигурацию для локальной разработки
func LocalConfig() *Config {
	return &Config{
		BaseURL:            "http://localhost:8080",
		Timeout:            30 * time.Second,
		UserAgent:          DefaultUserAgent,
		MaxRetries:         1,
		RetryDelay:         500 * time.Millisecond,
		InsecureSkipVerify: true,
	}
}

// ProductionConfig возвращает конфигурацию для продакшена
func ProductionConfig(baseURL string) *Config {
	return &Config{
		BaseURL:            baseURL,
		Timeout:            60 * time.Second,
		UserAgent:          DefaultUserAgent,
		MaxRetries:         5,
		RetryDelay:         2 * time.Second,
		InsecureSkipVerify: false,
	}
}

// TestingConfig возвращает конфигурацию для тестирования
func TestingConfig(baseURL string) *Config {
	return &Config{
		BaseURL:            baseURL,
		Timeout:            5 * time.Second,
		UserAgent:          "GophKeeper-Test/1.0",
		MaxRetries:         0,
		RetryDelay:         0,
		InsecureSkipVerify: true,
	}
}