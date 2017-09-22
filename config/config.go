package config

import (
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Listen          *string                   `yaml:"listen" validate:"required"`
	ExporterConfigs map[string]ExporterConfig `yaml:"exporters" validate:"required,dive"`
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

	c := &Config{}
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
