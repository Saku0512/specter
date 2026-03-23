package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RouteResponse struct {
	Status      int    `yaml:"status,omitempty"`
	ContentType string `yaml:"content_type,omitempty"`
	Response    any    `yaml:"response,omitempty"`
	File        string `yaml:"file,omitempty"`
}

type RouteMatch struct {
	Query       map[string]string `yaml:"query,omitempty"`
	Body        map[string]any    `yaml:"body,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	Status      int               `yaml:"status,omitempty"`
	ContentType string            `yaml:"content_type,omitempty"`
	Response    any               `yaml:"response,omitempty"`
	File        string            `yaml:"file,omitempty"`
}

type Webhook struct {
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method,omitempty"`  // default: POST
	Body    any               `yaml:"body,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Delay   int               `yaml:"delay,omitempty"` // milliseconds before sending
}

type Route struct {
	Path        string            `yaml:"path"`
	Method      string            `yaml:"method"`
	Status      int               `yaml:"status,omitempty"`
	Delay       int               `yaml:"delay,omitempty"` // milliseconds
	Headers     map[string]string `yaml:"headers,omitempty"`
	ContentType string            `yaml:"content_type,omitempty"`
	Response    any               `yaml:"response,omitempty"`
	Mode        string            `yaml:"mode,omitempty"` // "sequential" (default) or "random"
	Responses   []RouteResponse   `yaml:"responses,omitempty"`
	Match       []RouteMatch      `yaml:"match,omitempty"`
	RateLimit   int               `yaml:"rate_limit,omitempty"`  // max requests allowed
	RateReset   int               `yaml:"rate_reset,omitempty"`  // seconds until count resets
	State       string            `yaml:"state,omitempty"`       // required server state to match
	SetState    *string           `yaml:"set_state,omitempty"`   // state to set after responding
	Vars        map[string]string `yaml:"vars,omitempty"`        // require these vars to match
	SetVars     map[string]string `yaml:"set_vars,omitempty"`    // set these vars after responding
	Webhook     *Webhook          `yaml:"webhook,omitempty"`     // outgoing callback after responding
	File        string            `yaml:"file,omitempty"`        // path to response file (.json/.yaml/.yml)
	ErrorRate   float64           `yaml:"error_rate,omitempty"`  // 0.0-1.0 probability of injecting an error
	ErrorStatus int               `yaml:"error_status,omitempty"` // status code for injected error (default 503)
	DelayMin    int               `yaml:"delay_min,omitempty"`   // min random delay in ms (used with delay_max)
	DelayMax    int               `yaml:"delay_max,omitempty"`   // max random delay in ms
}

type Config struct {
	CORS    bool    `yaml:"cors,omitempty"`
	Proxy   string  `yaml:"proxy,omitempty"`
	OpenAPI string  `yaml:"openapi,omitempty"` // path to OpenAPI spec for request validation
	Routes  []Route `yaml:"routes"`
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
