package cli

import (
	"github.com/rrreeeyyy/exporter_proxy/types"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
)

type Config struct {
	Listen    *string                    `yaml:"listen" validate:"required"`
	Exporters map[string]ExportersConfig `yaml:"exporters" validate:"required,dive"`
}

type ExportersConfig struct {
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

func (c *Config) BuildExporters() ([]*types.Exporter, error) {
	exporters := []*types.Exporter{}
	for _, e := range c.Exporters {
		url, err := url.Parse(*e.URL)
		if err != nil {
			return nil, err
		}

		exporters = append(
			exporters,
			&types.Exporter{
				URL:  url,
				Path: e.Path,
			},
		)
	}

	return exporters, nil
}
