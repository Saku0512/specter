package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Saku0512/specter/config"
)

func TestRedactURL_redactsSensitiveQueryAndUserinfo(t *testing.T) {
	got := redactURL("https://user:pass@example.com/callback?token=abc123&name=alice&api_key=secret")
	if strings.Contains(got, "user") || strings.Contains(got, "pass") || strings.Contains(got, "abc123") || strings.Contains(got, "secret") {
		t.Fatalf("expected sensitive URL parts to be redacted, got %q", got)
	}
	if !strings.Contains(got, "name=alice") {
		t.Fatalf("expected non-sensitive query value to remain, got %q", got)
	}
}

func TestRedactBodyForLog_redactsSensitiveJSONFields(t *testing.T) {
	got := redactBodyForLog([]byte(`{"user":"alice","password":"pw","nested":{"token":"tok"},"items":[{"secret":"s"}]}`))
	for _, secret := range []string{`"pw"`, `"tok"`, `"s"`} {
		if strings.Contains(got, secret) {
			t.Fatalf("expected %q to be redacted from %q", secret, got)
		}
	}
	if !strings.Contains(got, `"user":"alice"`) {
		t.Fatalf("expected non-sensitive field to remain, got %q", got)
	}
}

func TestRedactBodyForLog_redactsFormFields(t *testing.T) {
	got := redactBodyForLog([]byte("username=alice&password=pw&token=tok"))
	values, err := url.ParseQuery(got)
	if err != nil {
		t.Fatalf("expected redacted form body to be parseable, got %q: %v", got, err)
	}
	if values.Get("password") == "pw" || values.Get("token") == "tok" {
		t.Fatalf("expected sensitive form values to be redacted, got %q", got)
	}
	if !strings.Contains(got, "username=alice") {
		t.Fatalf("expected non-sensitive field to remain, got %q", got)
	}
}

func TestVerboseLogger_redactsSensitiveHeadersQueryAndBody(t *testing.T) {
	var logs bytes.Buffer
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&logs)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
	}()

	srv := New(&config.Config{
		Routes: []config.Route{
			{Path: "/login", Method: "POST", Response: map[string]any{"ok": true}},
		},
	}, true, false)

	req := httptest.NewRequest(http.MethodPost, "/login?token=query-secret&name=alice", strings.NewReader(`{"password":"body-secret","user":"alice"}`))
	req.Header.Set("Authorization", "Bearer header-secret")
	req.Header.Set("X-Trace-Id", "trace-123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	got := logs.String()
	for _, secret := range []string{"query-secret", "header-secret", "body-secret"} {
		if strings.Contains(got, secret) {
			t.Fatalf("expected %q to be redacted from logs:\n%s", secret, got)
		}
	}
	for _, expected := range []string{"name=alice", "trace-123", `"user":"alice"`} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q to remain in logs:\n%s", expected, got)
		}
	}
}
