package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RouteResponse struct {
	Status      int    `yaml:"status,omitempty"`
	ContentType string `yaml:"content_type,omitempty"`
	Response    any    `yaml:"response,omitempty"`
	File        string `yaml:"file,omitempty"`
	OnCall      int    `yaml:"on_call,omitempty"` // match only on this call number (1-based)
	Script      string `yaml:"script,omitempty"`  // Go template producing the response body
}

type RouteMatch struct {
	Query       map[string]string `yaml:"query,omitempty"`
	Body        map[string]any    `yaml:"body,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	BodyPath    map[string]string `yaml:"body_path,omitempty"` // dot-notation path → regex pattern
	Status      int               `yaml:"status,omitempty"`
	ContentType string            `yaml:"content_type,omitempty"`
	Response    any               `yaml:"response,omitempty"`
	File        string            `yaml:"file,omitempty"`
	Script      string            `yaml:"script,omitempty"`   // Go template producing the response body
	Form        map[string]string `yaml:"form,omitempty"`      // match application/x-www-form-urlencoded fields (regex)
	GraphQL     *GraphQLMatch     `yaml:"graphql,omitempty"`   // match GraphQL operationName / variables
	SetState    *string           `yaml:"set_state,omitempty"` // transition server state after this match
	SetVars     map[string]string `yaml:"set_vars,omitempty"`  // set vars after this match
}

// GraphQLMatch selects a match entry by GraphQL operation name and/or variables.
type GraphQLMatch struct {
	Operation string            `yaml:"operation,omitempty"` // operationName regex/exact
	Variables map[string]string `yaml:"variables,omitempty"` // variable key → regex/exact
}

// StreamEvent is a single SSE event emitted by a streaming route.
type StreamEvent struct {
	Data  any    `yaml:"data,omitempty"`  // event payload (string or JSON-serialisable)
	Event string `yaml:"event,omitempty"` // SSE event type name (default: omitted → "message")
	ID    string `yaml:"id,omitempty"`    // SSE event ID
	Delay int    `yaml:"delay,omitempty"` // milliseconds to wait before sending this event
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
	OnCall      int               `yaml:"on_call,omitempty"`     // match only on this call number (1-based)
	Script      string            `yaml:"script,omitempty"`      // Go template producing the response body
	Proxy       string            `yaml:"proxy,omitempty"`       // forward this route to a real backend
	StorePush   string            `yaml:"store_push,omitempty"`  // push request body into named store → 201
	StoreList   string            `yaml:"store_list,omitempty"`  // list all items in named store → 200
	StoreGet    string            `yaml:"store_get,omitempty"`   // get item by store_key param → 200/404
	StorePut    string            `yaml:"store_put,omitempty"`   // replace/upsert item by store_key param → 200
	StorePatch  string            `yaml:"store_patch,omitempty"` // merge into item by store_key param → 200/404
	StoreDelete string            `yaml:"store_delete,omitempty"` // delete item by store_key param → 204/404
	StoreClear  string            `yaml:"store_clear,omitempty"`  // clear all items in named store → 204
	StoreKey     string        `yaml:"store_key,omitempty"`     // path param used as item ID (default: "id")
	Stream       bool          `yaml:"stream,omitempty"`        // respond with a Server-Sent Events stream
	Events       []StreamEvent `yaml:"events,omitempty"`        // SSE events to emit (requires stream: true)
	StreamRepeat bool          `yaml:"stream_repeat,omitempty"` // loop events until client disconnects
}

type Config struct {
	CORS          bool     `yaml:"cors,omitempty"`
	Proxy         string   `yaml:"proxy,omitempty"`
	OpenAPI       string   `yaml:"openapi,omitempty"`        // path to OpenAPI spec for request validation
	OpenAPIStrict bool     `yaml:"openapi_strict,omitempty"` // return 400 on validation failures instead of warning
	Include       []string `yaml:"include,omitempty"`        // glob patterns of additional config files to merge
	Routes        []Route  `yaml:"routes"`
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
	var resolved string
	for _, p := range candidates {
		data, err = os.ReadFile(p)
		if err == nil {
			resolved = p
			break
		}
	}
	if err != nil {
		return nil, err
	}

	absResolved, _ := filepath.Abs(resolved)
	seen := map[string]bool{absResolved: true}
	return loadData(data, filepath.Dir(resolved), seen)
}

// loadData unmarshals cfg from data and recursively merges any included files.
// dir is the base directory for resolving relative include patterns.
// seen prevents re-loading the same file and breaks cycles.
func loadData(data []byte, dir string, seen map[string]bool) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	for _, pattern := range cfg.Include {
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(dir, pattern)
		}
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("include %q: %w", pattern, err)
		}
		for _, match := range matches {
			abs, _ := filepath.Abs(match)
			if seen[abs] {
				continue
			}
			seen[abs] = true
			fileData, err := os.ReadFile(match)
			if err != nil {
				return nil, fmt.Errorf("include %q: %w", match, err)
			}
			sub, err := loadData(fileData, filepath.Dir(match), seen)
			if err != nil {
				return nil, fmt.Errorf("include %q: %w", match, err)
			}
			cfg.Routes = append(cfg.Routes, sub.Routes...)
		}
	}

	return &cfg, nil
}
