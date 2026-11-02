package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultShutDownTimeout = "10s"
)

type Config struct {
	Listen          *string                   `yaml:"listen" validate:"required"`
	TLSConfig       *TLSConfig                `yaml:"tls"`
	ShutDownTimeout *time.Duration            `yaml:"shutdown_timeout"`
	ExporterConfigs map[string]ExporterConfig `yaml:"exporters" validate:"required,min=1,dive"`
	AccessLogConfig *AccessLogConfig          `yaml:"access_log"`
	ErrorLogConfig  *ErrorLogConfig           `yaml:"error_log"`
}

type TLSConfig struct {
	CertFile *string `yaml:"certfile" validate:"required"`
	KeyFile  *string `yaml:"keyfile" validate:"required"`
}

type AccessLogConfig struct {
	Format *string  `yaml:"format" validate:"required"`
	Path   *string  `yaml:"path"`
	Fields []string `yaml:"fields" validate:"required"`
}

type ErrorLogConfig struct {
	Path *string `yaml:"path" validate:"required"`
}

type ExporterConfig struct {
	URL                *string `yaml:"url" validate:"required"`
	Path               *string `yaml:"path" validate:"required"`
	InsecureSkipVerify *bool   `yaml:"insecure_skip_verify"`
}

func LoadConfigFromYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sdt, _ := time.ParseDuration(DefaultShutDownTimeout)

	var raw struct {
		Listen          *string          `yaml:"listen"`
		TLSConfig       *TLSConfig       `yaml:"tls"`
		ShutDownTimeout *time.Duration   `yaml:"shutdown_timeout"`
		Exporters       exportersSection `yaml:"exporters"`
		AccessLogConfig *AccessLogConfig `yaml:"access_log"`
		ErrorLogConfig  *ErrorLogConfig  `yaml:"error_log"`
	}
	raw.ShutDownTimeout = &sdt

	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	err = raw.Exporters.Resolve(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	return &Config{
		Listen:          raw.Listen,
		TLSConfig:       raw.TLSConfig,
		ShutDownTimeout: raw.ShutDownTimeout,
		ExporterConfigs: raw.Exporters.items,
		AccessLogConfig: raw.AccessLogConfig,
		ErrorLogConfig:  raw.ErrorLogConfig,
	}, nil
}

func (c *Config) Validate() error {
	v := validator.New()
	err := v.Struct(c)
	return err
}
