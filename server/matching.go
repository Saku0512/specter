package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Saku0512/specter/config"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func matchesBody(body []byte, expected map[string]any) bool {
	if len(expected) == 0 {
		return true
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	for k, v := range expected {
		if fmt.Sprint(parsed[k]) != fmt.Sprint(v) {
			return false
		}
	}
	return true
}

func matchesHeaders(c *gin.Context, headers map[string]string) bool {
	for k, pattern := range headers {
		if !matchRegexOrExact(c.GetHeader(k), pattern) {
			return false
		}
	}
	return true
}

func matchesVars(vs *VarStore, expected map[string]string) bool {
	for k, v := range expected {
		if vs.Get(k) != v {
			return false
		}
	}
	return true
}

func matchesQuery(c *gin.Context, query map[string]string) bool {
	for k, pattern := range query {
		if !matchRegexOrExact(c.Query(k), pattern) {
			return false
		}
	}
	return true
}

// matchRegexOrExact treats pattern as a Go regular expression.
// If the pattern fails to compile it falls back to exact string comparison.
func matchRegexOrExact(actual, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return actual == pattern
	}
	return re.MatchString(actual)
}

// getJSONPath traverses a nested map using a dot-separated path.
// e.g. "user.role" returns the value at data["user"]["role"].
func getJSONPath(data map[string]any, path string) (string, bool) {
	idx := strings.Index(path, ".")
	if idx < 0 {
		v, ok := data[path]
		if !ok {
			return "", false
		}
		return fmt.Sprint(v), true
	}
	nested, ok := data[path[:idx]].(map[string]any)
	if !ok {
		return "", false
	}
	return getJSONPath(nested, path[idx+1:])
}

// matchesForm checks application/x-www-form-urlencoded fields against expected key→regex/exact patterns.
func matchesForm(c *gin.Context, body []byte, form map[string]string) bool {
	if len(form) == 0 {
		return true
	}
	if !strings.Contains(c.ContentType(), "application/x-www-form-urlencoded") {
		return false
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return false
	}
	for k, pattern := range form {
		if !matchRegexOrExact(values.Get(k), pattern) {
			return false
		}
	}
	return true
}

// matchesGraphQL matches a GraphQL request by operationName and/or variable values.
// Both fields support regex patterns (same as query/headers).
func matchesGraphQL(body []byte, gql *config.GraphQLMatch) bool {
	if gql == nil {
		return true
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	if gql.Operation != "" {
		op, _ := payload["operationName"].(string)
		if !matchRegexOrExact(op, gql.Operation) {
			return false
		}
	}
	if len(gql.Variables) > 0 {
		vars, _ := payload["variables"].(map[string]any)
		for k, pattern := range gql.Variables {
			if !matchRegexOrExact(fmt.Sprint(vars[k]), pattern) {
				return false
			}
		}
	}
	return true
}

// matchesBodyPath checks dot-notation path → regex pattern conditions against the request body.
func matchesBodyPath(body []byte, paths map[string]string) bool {
	if len(paths) == 0 {
		return true
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	for path, pattern := range paths {
		actual, ok := getJSONPath(parsed, path)
		if !ok {
			return false
		}
		re, err := regexp.Compile(pattern)
		if err != nil || !re.MatchString(actual) {
			return false
		}
	}
	return true
}

// matchesCookies checks request cookies against expected name→regex/exact patterns.
func matchesCookies(c *gin.Context, cookies map[string]string) bool {
	if len(cookies) == 0 {
		return true
	}
	for name, pattern := range cookies {
		cookie, err := c.Cookie(name)
		if err != nil {
			return false
		}
		re, err := regexp.Compile(pattern)
		if err != nil || !re.MatchString(cookie) {
			return false
		}
	}
	return true
}

// applySetCookies writes Set-Cookie headers for each cookie in the list.
func applySetCookies(c *gin.Context, cookies []config.SetCookie) {
	for _, sc := range cookies {
		sameSite := http.SameSiteDefaultMode
		switch sc.SameSite {
		case "Strict", "strict":
			sameSite = http.SameSiteStrictMode
		case "Lax", "lax":
			sameSite = http.SameSiteLaxMode
		case "None", "none":
			sameSite = http.SameSiteNoneMode
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     sc.Name,
			Value:    sc.Value,
			Path:     sc.Path,
			Domain:   sc.Domain,
			MaxAge:   sc.MaxAge,
			HttpOnly: sc.HTTPOnly,
			Secure:   sc.Secure,
			SameSite: sameSite,
		})
	}
}

// matchesBodySchema validates the request body against an inline JSON Schema.
// The schema is expressed as a map[string]any (same structure as an OpenAPI schema object).
func matchesBodySchema(body []byte, schema map[string]any) bool {
	if len(schema) == 0 {
		return true
	}
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return false
	}
	sc := &openapi3.Schema{}
	if err = json.Unmarshal(schemaBytes, sc); err != nil {
		return false
	}
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return false
	}
	return sc.VisitJSON(value) == nil
}
