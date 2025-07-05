package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "GophKeeper-Client/1.0", config.UserAgent)
	assert.Equal(t, 3, config.MaxRetries)
	assert.False(t, config.InsecureSkipVerify)
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{
			name:    "valid HTTP URL",
			baseURL: "http://localhost:8080",
		},
		{
			name:    "valid HTTPS URL",
			baseURL: "https://api.example.com",
		},
		{
			name:    "URL with path",
			baseURL: "https://api.example.com/v1",
		},
		{
			name:    "empty URL",
			baseURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig(tt.baseURL)
			assert.NotNil(t, config)
			assert.Equal(t, tt.baseURL, config.BaseURL)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				UserAgent: "Test-Client",
			},
			wantErr: false,
		},
		{
			name: "empty base URL",
			config: &Config{
				BaseURL:   "",
				Timeout:   30 * time.Second,
				UserAgent: "Test-Client",
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: &Config{
				BaseURL:   "not-a-url",
				Timeout:   30 * time.Second,
				UserAgent: "Test-Client",
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				BaseURL:   "https://api.example.com",
				Timeout:   0,
				UserAgent: "Test-Client",
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &Config{
				BaseURL:   "https://api.example.com",
				Timeout:   -5 * time.Second,
				UserAgent: "Test-Client",
			},
			wantErr: true,
		},
		{
			name: "empty user agent",
			config: &Config{
				BaseURL:   "https://api.example.com",
				Timeout:   30 * time.Second,
				UserAgent: "",
			},
			wantErr: true,
		},
		{
			name: "negative max retries",
			config: &Config{
				BaseURL:    "https://api.example.com",
				Timeout:    30 * time.Second,
				UserAgent:  "Test-Client",
				MaxRetries: -1,
			},
			wantErr: true,
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

func TestConfig_String(t *testing.T) {
	config := &Config{
		BaseURL:            "https://api.example.com",
		Timeout:            30 * time.Second,
		UserAgent:          "Test-Client",
		MaxRetries:         3,
		InsecureSkipVerify: false,
	}

	str := config.String()
	assert.Contains(t, str, "https://api.example.com")
	assert.Contains(t, str, "30s")
	assert.Contains(t, str, "Test-Client")
	assert.Contains(t, str, "3")
	assert.Contains(t, str, "false")
}

func TestConfig_ChainedMethods(t *testing.T) {
	config := DefaultConfig().
		WithTimeout(60*time.Second).
		WithUserAgent("Custom/1.0").
		WithRetry(10, 3*time.Second).
		WithInsecureSkipVerify(true)

	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.Equal(t, "Custom/1.0", config.UserAgent)
	assert.Equal(t, 10, config.MaxRetries)
	assert.True(t, config.InsecureSkipVerify)
}
