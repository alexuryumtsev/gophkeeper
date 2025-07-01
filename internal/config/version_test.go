package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultVersion(t *testing.T) {
	version := GetDefaultVersion()

	assert.NotNil(t, version)
	assert.Equal(t, "dev", version.Version)
	assert.Equal(t, "unknown", version.BuildDate)
	assert.Equal(t, "unknown", version.GitCommit)
}

func TestSetVersion(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		buildDate    string
		gitCommit    string
		expectedVer  string
		expectedDate string
		expectedGit  string
	}{
		{
			name:         "set all values",
			version:      "v1.0.0",
			buildDate:    "2023-12-01",
			gitCommit:    "abc123",
			expectedVer:  "v1.0.0",
			expectedDate: "2023-12-01",
			expectedGit:  "abc123",
		},
		{
			name:         "set only version",
			version:      "v2.0.0",
			buildDate:    "",
			gitCommit:    "",
			expectedVer:  "v2.0.0",
			expectedDate: "unknown", // не изменился
			expectedGit:  "unknown", // не изменился
		},
		{
			name:         "set only build date",
			version:      "",
			buildDate:    "2023-12-25",
			gitCommit:    "",
			expectedVer:  "dev",      // не изменился
			expectedDate: "2023-12-25",
			expectedGit:  "unknown", // не изменился
		},
		{
			name:         "set only git commit",
			version:      "",
			buildDate:    "",
			gitCommit:    "def456",
			expectedVer:  "dev",     // не изменился
			expectedDate: "unknown", // не изменился
			expectedGit:  "def456",
		},
		{
			name:         "empty values don't override",
			version:      "",
			buildDate:    "",
			gitCommit:    "",
			expectedVer:  "dev",
			expectedDate: "unknown",
			expectedGit:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новую версию для каждого теста
			version := GetDefaultVersion()

			version.SetVersion(tt.version, tt.buildDate, tt.gitCommit)

			assert.Equal(t, tt.expectedVer, version.Version)
			assert.Equal(t, tt.expectedDate, version.BuildDate)
			assert.Equal(t, tt.expectedGit, version.GitCommit)
		})
	}
}

func TestVersionStruct(t *testing.T) {
	// Тест создания структуры напрямую
	version := &Version{
		Version:   "v1.5.0",
		BuildDate: "2023-11-15",
		GitCommit: "xyz789",
	}

	assert.Equal(t, "v1.5.0", version.Version)
	assert.Equal(t, "2023-11-15", version.BuildDate)
	assert.Equal(t, "xyz789", version.GitCommit)
}

func TestSetVersionPartial(t *testing.T) {
	version := GetDefaultVersion()

	// Сначала устанавливаем все значения
	version.SetVersion("v1.0.0", "2023-01-01", "abc123")

	// Теперь обновляем только версию
	version.SetVersion("v1.1.0", "", "")

	assert.Equal(t, "v1.1.0", version.Version)
	assert.Equal(t, "2023-01-01", version.BuildDate) // не изменился
	assert.Equal(t, "abc123", version.GitCommit)     // не изменился

	// Обновляем только дату сборки
	version.SetVersion("", "2023-02-01", "")

	assert.Equal(t, "v1.1.0", version.Version)       // не изменился
	assert.Equal(t, "2023-02-01", version.BuildDate)
	assert.Equal(t, "abc123", version.GitCommit) // не изменился

	// Обновляем только git commit
	version.SetVersion("", "", "def456")

	assert.Equal(t, "v1.1.0", version.Version)       // не изменился
	assert.Equal(t, "2023-02-01", version.BuildDate) // не изменился
	assert.Equal(t, "def456", version.GitCommit)
}