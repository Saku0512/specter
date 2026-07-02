package server

import (
	"testing"

	"github.com/Saku0512/specter/config"
)

func FuzzMatchesBodyPath(f *testing.F) {
	seeds := []struct {
		body    string
		path    string
		pattern string
	}{
		{`{"status":"pending","user":{"role":"admin"}}`, "user.role", "^admin$"},
		{`{"order":{"total":42,"status":"submitted"}}`, "order.total", "^42$"},
		{`{"items":[{"sku":"abc"}]}`, "items", "sku"},
		{`not-json`, "status", ".*"},
		{`{"status":"pending"}`, "status", "["},
	}
	for _, seed := range seeds {
		f.Add(seed.body, seed.path, seed.pattern)
	}

	f.Fuzz(func(t *testing.T, body, path, pattern string) {
		if len(body) > 64*1024 || len(path) > 1024 || len(pattern) > 4096 {
			return
		}
		if !matchesBodyPath([]byte(body), nil) {
			t.Fatal("nil body_path conditions should always match")
		}
		_ = matchesBodyPath([]byte(body), map[string]string{path: pattern})
	})
}

func FuzzMatchesGraphQL(f *testing.F) {
	seeds := []struct {
		body       string
		operation  string
		variable   string
		varPattern string
	}{
		{`{"operationName":"GetUser","variables":{"id":"42"}}`, "^GetUser$", "id", "^42$"},
		{`{"operationName":"CreateUser","variables":{"role":"admin"}}`, "Create", "role", "admin"},
		{`{"query":"query { viewer { id } }"}`, "", "", ""},
		{`not-json`, ".*", "id", ".*"},
		{`{"operationName":"DeleteUser","variables":{"id":"abc"}}`, "[", "id", "^abc$"},
	}
	for _, seed := range seeds {
		f.Add(seed.body, seed.operation, seed.variable, seed.varPattern)
	}

	f.Fuzz(func(t *testing.T, body, operation, variable, varPattern string) {
		if len(body) > 64*1024 || len(operation) > 4096 || len(variable) > 1024 || len(varPattern) > 4096 {
			return
		}
		if !matchesGraphQL([]byte(body), nil) {
			t.Fatal("nil GraphQL matcher should always match")
		}
		gql := &config.GraphQLMatch{Operation: operation}
		if variable != "" || varPattern != "" {
			gql.Variables = map[string]string{variable: varPattern}
		}
		_ = matchesGraphQL([]byte(body), gql)
	})
}
