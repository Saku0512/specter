package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// --- convertPath ---

func TestConvertPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/users/{id}", "/users/:id"},
		{"/a/{b}/c/{d}", "/a/:b/c/:d"},
		{"/no-params", "/no-params"},
	}
	for _, c := range cases {
		got := convertPath(c.in)
		if got != c.want {
			t.Errorf("convertPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- schemaToExample ---

func strType() *openapi3.Types  { t := openapi3.Types{"string"}; return &t }
func intType() *openapi3.Types  { t := openapi3.Types{"integer"}; return &t }
func numType() *openapi3.Types  { t := openapi3.Types{"number"}; return &t }
func boolType() *openapi3.Types { t := openapi3.Types{"boolean"}; return &t }
func arrType() *openapi3.Types  { t := openapi3.Types{"array"}; return &t }
func objType() *openapi3.Types  { t := openapi3.Types{"object"}; return &t }

func TestSchemaToExample_string(t *testing.T) {
	got := schemaToExample(&openapi3.Schema{Type: strType()})
	if got != "string" {
		t.Errorf("expected string, got %v", got)
	}
}

func TestSchemaToExample_integer(t *testing.T) {
	got := schemaToExample(&openapi3.Schema{Type: intType()})
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestSchemaToExample_number(t *testing.T) {
	got := schemaToExample(&openapi3.Schema{Type: numType()})
	if got != 0.0 {
		t.Errorf("expected 0.0, got %v", got)
	}
}

func TestSchemaToExample_boolean(t *testing.T) {
	got := schemaToExample(&openapi3.Schema{Type: boolType()})
	if got != true {
		t.Errorf("expected true, got %v", got)
	}
}

func TestSchemaToExample_array(t *testing.T) {
	schema := &openapi3.Schema{
		Type:  arrType(),
		Items: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: strType()}},
	}
	got, ok := schemaToExample(schema).([]any)
	if !ok || len(got) != 1 {
		t.Fatalf("expected []any{...}, got %v", got)
	}
	if got[0] != "string" {
		t.Errorf("expected string item, got %v", got[0])
	}
}

func TestSchemaToExample_object(t *testing.T) {
	schema := &openapi3.Schema{
		Type: objType(),
		Properties: openapi3.Schemas{
			"name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: strType()}},
			"age":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: intType()}},
		},
	}
	got, ok := schemaToExample(schema).(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", got)
	}
	if got["name"] != "string" {
		t.Errorf("expected name:string, got %v", got["name"])
	}
	if got["age"] != 0 {
		t.Errorf("expected age:0, got %v", got["age"])
	}
}

func TestSchemaToExample_useExampleField(t *testing.T) {
	schema := &openapi3.Schema{
		Type:    strType(),
		Example: "hello",
	}
	got := schemaToExample(schema)
	if got != "hello" {
		t.Errorf("expected hello, got %v", got)
	}
}

// --- Full generation ---

const sampleSpec = `
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          content:
            application/json:
              example:
                - id: 1
                  name: Alice
    post:
      responses:
        '201':
          content:
            application/json:
              example:
                message: created
  /users/{id}:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
`

func TestRun(t *testing.T) {
	dir := t.TempDir()
	inFile := filepath.Join(dir, "openapi.yml")
	outFile := filepath.Join(dir, "config.yml")

	if err := os.WriteFile(inFile, []byte(sampleSpec), 0o644); err != nil {
		t.Fatal(err)
	}

	Run([]string{"-i", inFile, "-o", outFile})

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	out := string(data)

	for _, want := range []string{"/users", "/users/:id", "GET", "POST", "201"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output:\n%s", want, out)
		}
	}
}
