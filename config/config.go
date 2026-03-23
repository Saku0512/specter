package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RouteResponse struct {
	Status   int `yaml:"status,omitempty"`
	Response any `yaml:"response,omitempty"`
}

type Route struct {
	Path      string            `yaml:"path"`
	Method    string            `yaml:"method"`
	Status    int               `yaml:"status,omitempty"`
	Delay     int               `yaml:"delay,omitempty"` // milliseconds
	Headers   map[string]string `yaml:"headers,omitempty"`
	Response  any               `yaml:"response,omitempty"`
	Mode      string            `yaml:"mode,omitempty"` // "sequential" (default) or "random"
	Responses []RouteResponse   `yaml:"responses,omitempty"`
}

type Config struct {
	CORS   bool    `yaml:"cors,omitempty"`
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
