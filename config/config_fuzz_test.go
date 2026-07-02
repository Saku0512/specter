package config

import "testing"

func FuzzLoadBytes(f *testing.F) {
	seeds := [][]byte{
		[]byte("routes: []\n"),
		[]byte("routes:\n  - path: /users\n    method: GET\n    response:\n      ok: true\n"),
		[]byte("cors: true\ninclude:\n  - extra.yml\nroutes:\n  - path: /graphql\n    method: POST\n    match:\n      - graphql:\n          operation: GetUser\n        response: ok\n"),
		[]byte("routes:\n  - path: /orders\n    method: POST\n    match:\n      - body_path:\n          order.status: '^submitted$'\n        status: 201\n"),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		cfg, err := LoadBytes(data)
		if err != nil {
			return
		}
		if cfg == nil {
			t.Fatal("LoadBytes returned nil config without an error")
		}
	})
}
