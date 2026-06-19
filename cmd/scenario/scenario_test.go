package scenario

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchScenarios(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/__specter/scenarios" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"active":"login-success","scenarios":["login-success","logged-out"]}`))
	}))
	defer ts.Close()

	got, err := fetchScenarios(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if got.Active != "login-success" {
		t.Errorf("expected active login-success, got %q", got.Active)
	}
	if len(got.Scenarios) != 2 || got.Scenarios[0] != "login-success" || got.Scenarios[1] != "logged-out" {
		t.Errorf("unexpected scenarios: %v", got.Scenarios)
	}
}

func TestApplyScenario(t *testing.T) {
	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/__specter/scenarios/login-success" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		called = true
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"active":"login-success"}`))
	}))
	defer ts.Close()

	if err := applyScenario(ts.URL, "login-success"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected scenario endpoint to be called")
	}
}

func TestApplyScenario_notFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"scenario not found"}`, http.StatusNotFound)
	}))
	defer ts.Close()

	if err := applyScenario(ts.URL, "missing"); err == nil {
		t.Fatal("expected error")
	}
}
