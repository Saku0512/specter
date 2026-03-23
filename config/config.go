package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Route struct {
	Path		string	`yaml:"path"`
	Method		string 	`yaml:"method"`
	Status		int		`yaml:"status"`
	Response	any		`yaml:"response"`
}

type Config struct {
	Route []Route `yaml:"route"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
