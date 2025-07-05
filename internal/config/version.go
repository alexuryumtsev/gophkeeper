package config

// Version информация о версии приложения
type Version struct {
	Version   string
	BuildDate string
	GitCommit string
}

// GetDefaultVersion возвращает информацию о версии по умолчанию
func GetDefaultVersion() *Version {
	return &Version{
		Version:   "dev",
		BuildDate: "unknown",
		GitCommit: "unknown",
	}
}

// SetVersion устанавливает информацию о версии
func (v *Version) SetVersion(version, buildDate, gitCommit string) {
	if version != "" {
		v.Version = version
	}
	if buildDate != "" {
		v.BuildDate = buildDate
	}
	if gitCommit != "" {
		v.GitCommit = gitCommit
	}
}