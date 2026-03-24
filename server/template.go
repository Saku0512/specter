package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// applyParams recursively replaces ":paramName" strings in the response with
// actual path parameter values. Numeric params are converted to int or float64.
func applyParams(v any, params gin.Params) any {
	switch val := v.(type) {
	case string:
		if strings.HasPrefix(val, ":") {
			if p := params.ByName(val[1:]); p != "" {
				if n, err := strconv.Atoi(p); err == nil {
					return n
				}
				if f, err := strconv.ParseFloat(p, 64); err == nil {
					return f
				}
				return p
			}
		}
		return val
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = applyParams(v2, params)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = applyParams(v2, params)
		}
		return out
	default:
		return v
	}
}

func buildTemplateData(c *gin.Context, bodyBytes []byte) map[string]any {
	params := map[string]any{}
	for _, p := range c.Params {
		params[p.Key] = p.Value
	}
	query := map[string]any{}
	for k, vs := range c.Request.URL.Query() {
		if len(vs) == 1 {
			query[k] = vs[0]
		} else {
			query[k] = vs
		}
	}
	body := map[string]any{}
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &body)
	}
	headers := map[string]any{}
	for k, vs := range c.Request.Header {
		headers[k] = strings.Join(vs, ", ")
	}
	return map[string]any{
		"params":  params,
		"query":   query,
		"body":    body,
		"headers": headers,
		"method":  c.Request.Method,
		"path":    c.Request.URL.Path,
	}
}

// buildFuncMap returns the template FuncMap, including store-aware functions.
// store may be nil (template calls will return empty results in that case).
func buildFuncMap(store *DataStore) template.FuncMap {
	fm := template.FuncMap{
		"store": func(name string) []map[string]any {
			if store == nil {
				return nil
			}
			return store.List(name)
		},
		"storeGet": func(name, id string) map[string]any {
			if store == nil {
				return nil
			}
			item, _ := store.Get(name, id)
			return item
		},
		"storeCount": func(name string) int {
			if store == nil {
				return 0
			}
			return len(store.List(name))
		},
		"fake": func(kind string) string {
		switch kind {
		case "name":
			return gofakeit.Name()
		case "first_name":
			return gofakeit.FirstName()
		case "last_name":
			return gofakeit.LastName()
		case "email":
			return gofakeit.Email()
		case "uuid":
			return gofakeit.UUID()
		case "phone":
			return gofakeit.Phone()
		case "url":
			return gofakeit.URL()
		case "ip":
			return gofakeit.IPv4Address()
		case "username":
			return gofakeit.Username()
		case "password":
			return gofakeit.Password(true, true, true, false, false, 12)
		case "word":
			return gofakeit.Word()
		case "sentence":
			return gofakeit.Sentence(6)
		case "paragraph":
			return gofakeit.Paragraph(1, 3, 10, " ")
		case "color":
			return gofakeit.Color()
		case "country":
			return gofakeit.Country()
		case "city":
			return gofakeit.City()
		case "zip":
			return gofakeit.Zip()
		case "street":
			return gofakeit.Street()
		case "company":
			return gofakeit.Company()
		case "job":
			return gofakeit.JobTitle()
		case "int":
			return strconv.Itoa(gofakeit.IntRange(1, 10000))
		case "float":
			return strconv.FormatFloat(float64(gofakeit.Float32Range(0, 1000)), 'f', 2, 64)
		case "bool":
			return strconv.FormatBool(gofakeit.Bool())
		case "date":
			return gofakeit.Date().Format("2006-01-02")
		case "datetime":
			return gofakeit.Date().Format(time.RFC3339)
		default:
			return ""
		}
	},
	"default": func(def, val string) string {
		if val == "" {
			return def
		}
		return val
	},
	"upper":  strings.ToUpper,
	"lower":  strings.ToLower,
	"trim":   strings.TrimSpace,
	"now":    func() string { return time.Now().UTC().Format(time.RFC3339) },
	"add":    func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"json": func(v any) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	}
	return fm
}

// execScript renders a Go template script and returns the result plus an optional
// HTTP status override. If the JSON output contains a "_status" key (integer 100-599)
// it is extracted and returned as the status override; the key is removed from the body.
// If the output is valid JSON it is decoded; otherwise the raw string is returned.
func execScript(script string, data map[string]any, store *DataStore) (body any, statusOverride int) {
	if script == "" {
		return nil, 0
	}
	tmpl, err := template.New("").Funcs(buildFuncMap(store)).Parse(script)
	if err != nil {
		log.Printf("script: parse error: %v", err)
		return nil, 0
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("script: execute error: %v", err)
		return nil, 0
	}
	out := strings.TrimSpace(buf.String())
	var v any
	if err := json.Unmarshal([]byte(out), &v); err == nil {
		if m, ok := v.(map[string]any); ok {
			if s, ok := m["_status"].(float64); ok && s >= 100 && s < 600 {
				delete(m, "_status")
				return m, int(s)
			}
		}
		return v, 0
	}
	return out, 0
}

func applyTemplate(v any, data map[string]any, store *DataStore) any {
	switch val := v.(type) {
	case string:
		if !strings.Contains(val, "{{") {
			return val
		}
		tmpl, err := template.New("").Funcs(buildFuncMap(store)).Parse(val)
		if err != nil {
			return val
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return val
		}
		return buf.String()
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = applyTemplate(v2, data, store)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = applyTemplate(v2, data, store)
		}
		return out
	default:
		return v
	}
}

// loadFile reads a file and returns its parsed content and an inferred content type.
// .json files are parsed as JSON, .yaml/.yml as YAML (served as JSON-compatible data).
// All other files are returned as a raw string with an empty content type.
func loadFile(path string) (any, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", path, err)
		}
		return v, "application/json", nil
	case ".yaml", ".yml":
		var v any
		if err := yaml.Unmarshal(data, &v); err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", path, err)
		}
		return v, "application/json", nil
	default:
		return string(data), "", nil
	}
}

// resolveBody returns the response body, an inferred content type, and an optional
// HTTP status override (non-zero only when script uses the _status envelope).
// Priority: script > file > body.
func resolveBody(body any, file, script string, tmplData map[string]any, params gin.Params, store *DataStore) (any, string, int) {
	if script != "" {
		b, statusOverride := execScript(script, tmplData, store)
		return b, "", statusOverride
	}
	if file != "" {
		data, ct, err := loadFile(file)
		if err != nil {
			log.Printf("file response: %v", err)
			return map[string]any{"error": "failed to load response file"}, "", 0
		}
		return applyParams(applyTemplate(data, tmplData, store), params), ct, 0
	}
	return applyParams(applyTemplate(body, tmplData, store), params), "", 0
}

func respond(c *gin.Context, status int, contentType string, body any) {
	if contentType == "" || contentType == "application/json" {
		c.JSON(status, body)
		return
	}
	s, _ := body.(string)
	c.Data(status, contentType, []byte(s))
}

