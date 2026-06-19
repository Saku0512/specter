package config

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPublishedSchemaMetadata(t *testing.T) {
	schema := loadPublishedSchema(t)

	if got := schema["$id"]; got != "https://specter.dev/schemas/config.schema.json" {
		t.Fatalf("unexpected schema id: %v", got)
	}
	if got := schema["x-specter-schema-version"]; got != "1" {
		t.Fatalf("unexpected schema version: %v", got)
	}
	if got := schema["x-specter-config-compatibility"]; got != "v0" {
		t.Fatalf("unexpected config compatibility: %v", got)
	}
}

func TestPublishedSchemaCoversConfigYAMLTags(t *testing.T) {
	schema := loadPublishedSchema(t)
	defs := objectAt(t, schema, "$defs")

	assertSchemaPropertiesMatchType(t, objectAt(t, schema, "properties"), reflect.TypeOf(Config{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "route"), reflect.TypeOf(Route{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "routeMatch"), reflect.TypeOf(RouteMatch{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "routeResponse"), reflect.TypeOf(RouteResponse{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "graphqlMatch"), reflect.TypeOf(GraphQLMatch{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "streamEvent"), reflect.TypeOf(StreamEvent{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "setCookie"), reflect.TypeOf(SetCookie{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "webhook"), reflect.TypeOf(Webhook{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "latencyProfile"), reflect.TypeOf(LatencyProfile{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "scenario"), reflect.TypeOf(Scenario{}))
	assertSchemaPropertiesMatchType(t, propertiesForDef(t, defs, "store"), reflect.TypeOf(StoreConfig{}))
}

func loadPublishedSchema(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile("../schemas/config.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

func propertiesForDef(t *testing.T, defs map[string]any, name string) map[string]any {
	t.Helper()
	def := objectAt(t, defs, name)
	return objectAt(t, def, "properties")
}

func assertSchemaPropertiesMatchType(t *testing.T, properties map[string]any, typ reflect.Type) {
	t.Helper()
	want := yamlFields(typ)
	got := mapKeys(properties)

	for _, field := range want {
		if !containsString(got, field) {
			t.Fatalf("%s schema is missing yaml field %q; got %v", typ.Name(), field, got)
		}
	}
	for _, field := range got {
		if !containsString(want, field) {
			t.Fatalf("%s schema has unknown property %q; yaml fields are %v", typ.Name(), field, want)
		}
	}
}

func yamlFields(typ reflect.Type) []string {
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("yaml")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		fields = append(fields, name)
	}
	return fields
}

func objectAt(t *testing.T, obj map[string]any, key string) map[string]any {
	t.Helper()
	raw, ok := obj[key]
	if !ok {
		t.Fatalf("missing object key %q", key)
	}
	out, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("object key %q has type %T", key, raw)
	}
	return out
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
