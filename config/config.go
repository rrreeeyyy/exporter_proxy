package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

const (
	DefaultShutDownTimeout = "10s"
)

type Config struct {
	Listen          *string                   `yaml:"listen" validate:"required"`
	TLSConfig       *TLSConfig                `yaml:"tls"`
	ShutDownTimeout *time.Duration            `yaml:"shutdown_timeout"`
	ExporterConfigs map[string]ExporterConfig `yaml:"exporters" validate:"required,dive"`
	AccessLogConfig *AccessLogConfig          `yaml:"access_log"`
	ErrorLogConfig  *ErrorLogConfig           `yaml:"error_log"`
}

type TLSConfig struct {
	CertFile *string `yaml:"certfile"`
	KeyFile  *string `yaml:"keyfile"`
}

type AccessLogConfig struct {
	Format *string  `yaml:"format" validate:"required"`
	Path   *string  `yaml:"path"`
	Fields []string `yaml:"fields" validate:"required"`
}

type ErrorLogConfig struct {
	Path *string `yaml:"path" validate"required"`
}

type ExporterConfig struct {
	URL  *string `yaml:"url" validate:"required"`
	Path *string `yaml:"path" validate:"required"`
}

func LoadConfigFromYAML(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sdt, _ := time.ParseDuration(DefaultShutDownTimeout)

	c := &Config{ShutDownTimeout: &sdt}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Validate() error {
	v := validator.New()
	err := v.Struct(c)
	return err
}
