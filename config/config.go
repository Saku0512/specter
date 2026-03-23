package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Route struct {
	Path     string            `yaml:"path"`
	Method   string            `yaml:"method"`
	Status   int               `yaml:"status"`
	Delay    int               `yaml:"delay"` // milliseconds
	Headers  map[string]string `yaml:"headers"`
	Response any               `yaml:"response"`
}

type Config struct {
	CORS   bool    `yaml:"cors"`
	Routes []Route `yaml:"routes"`
}

func Load(path string) (*Config, error) {
	candidates := []string{path}
	if path == "config.yaml" {
		candidates = append(candidates, "config.yml")
	} else if path == "config.yml" {
		candidates = append(candidates, "config.yaml")
	}

	var data []byte
	var err error
	for _, p := range candidates {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
