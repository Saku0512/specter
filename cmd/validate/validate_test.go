package validate

import (
	"testing"

	"github.com/Saku0512/specter/config"
)

func TestCheck_valid(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{Path: "/users", Method: "GET", Status: 200},
			{Path: "/users", Method: "POST", Status: 201},
		},
	}
	if errs := check(cfg); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheck_missingPath(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Method: "GET"}}}
	errs := check(cfg)
	assertContains(t, errs, "missing path")
}

func TestCheck_missingMethod(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a"}}}
	errs := check(cfg)
	assertContains(t, errs, "missing method")
}

func TestCheck_invalidMethod(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GETS"}}}
	errs := check(cfg)
	assertContains(t, errs, "invalid method")
}

func TestCheck_invalidStatus(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", Status: 999}}}
	errs := check(cfg)
	assertContains(t, errs, "invalid status")
}

func TestCheck_invalidMode(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", Mode: "cycle"}}}
	errs := check(cfg)
	assertContains(t, errs, "invalid mode")
}

func TestCheck_validModes(t *testing.T) {
	for _, mode := range []string{"", "sequential", "random"} {
		cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", Mode: mode}}}
		if errs := check(cfg); len(errs) != 0 {
			t.Errorf("mode %q: expected no errors, got %v", mode, errs)
		}
	}
}

func TestCheck_matchNoCondition(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{
				Path:   "/a",
				Method: "GET",
				Match:  []config.RouteMatch{{Status: 200, Response: "ok"}},
			},
		},
	}
	errs := check(cfg)
	assertContains(t, errs, "must have at least one query, body, or headers condition")
}

func TestCheck_matchHeadersOnly(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{
				Path:   "/a",
				Method: "GET",
				Match:  []config.RouteMatch{{Headers: map[string]string{"Authorization": "Bearer token"}}},
			},
		},
	}
	if errs := check(cfg); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheck_matchQueryOnly(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{
				Path:   "/a",
				Method: "GET",
				Match:  []config.RouteMatch{{Query: map[string]string{"q": "1"}}},
			},
		},
	}
	if errs := check(cfg); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheck_rateLimitNegative(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", RateLimit: -1}}}
	errs := check(cfg)
	assertContains(t, errs, "rate_limit must be non-negative")
}

func TestCheck_rateResetWithoutLimit(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", RateReset: 60}}}
	errs := check(cfg)
	assertContains(t, errs, "rate_reset requires rate_limit")
}

func TestCheck_responsesInvalidStatus(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{
				Path:   "/a",
				Method: "GET",
				Responses: []config.RouteResponse{
					{Status: 999, Response: "x"},
				},
			},
		},
	}
	errs := check(cfg)
	assertContains(t, errs, "responses[0] invalid status")
}

func TestCheck_errorRateOutOfRange(t *testing.T) {
	for _, rate := range []float64{-0.1, 1.1} {
		cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", ErrorRate: rate}}}
		assertContains(t, check(cfg), "error_rate must be between")
	}
}

func TestCheck_errorRateValid(t *testing.T) {
	for _, rate := range []float64{0, 0.5, 1.0} {
		cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", ErrorRate: rate}}}
		if errs := check(cfg); len(errs) != 0 {
			t.Errorf("rate %v: expected no errors, got %v", rate, errs)
		}
	}
}

func TestCheck_errorStatusInvalid(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", ErrorStatus: 999}}}
	assertContains(t, check(cfg), "error_status invalid status")
}

func TestCheck_delayMinGtMax(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", DelayMin: 500, DelayMax: 100}}}
	assertContains(t, check(cfg), "delay_min must be <= delay_max")
}

func assertContains(t *testing.T, errs []string, substr string) {
	t.Helper()
	for _, e := range errs {
		if contains(e, substr) {
			return
		}
	}
	t.Errorf("expected error containing %q, got %v", substr, errs)
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}

func TestCheck_bodyPathInvalidRegex(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{
		Path:   "/a",
		Method: "GET",
		Match:  []config.RouteMatch{{BodyPath: map[string]string{"role": "["}}},
	}}}
	assertContains(t, check(cfg), "invalid regex")
}

func TestCheck_bodyPathOnlyConditionValid(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{
		Path:   "/a",
		Method: "GET",
		Match:  []config.RouteMatch{{BodyPath: map[string]string{"role": "^admin$"}}},
	}}}
	if errs := check(cfg); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestCheck_onCallNegative(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{Path: "/a", Method: "GET", OnCall: -1}}}
	assertContains(t, check(cfg), "on_call must be non-negative")
}

func TestCheck_responsesOnCallNegative(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{
		Path:      "/a",
		Method:    "GET",
		Responses: []config.RouteResponse{{OnCall: -1, Response: "x"}},
	}}}
	assertContains(t, check(cfg), "responses[0] on_call must be non-negative")
}

func TestCheck_queryInvalidRegex(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{
		Path:   "/a",
		Method: "GET",
		Match:  []config.RouteMatch{{Query: map[string]string{"q": "["}}},
	}}}
	assertContains(t, check(cfg), "query[\"q\"] invalid regex")
}

func TestCheck_headersInvalidRegex(t *testing.T) {
	cfg := &config.Config{Routes: []config.Route{{
		Path:   "/a",
		Method: "GET",
		Match:  []config.RouteMatch{{Headers: map[string]string{"Authorization": "["}}},
	}}}
	assertContains(t, check(cfg), "headers[\"Authorization\"] invalid regex")
}
