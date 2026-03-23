package gen

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Saku0512/specter/config"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func Run(args []string) {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	input := fs.String("i", "", "path to OpenAPI spec (YAML or JSON)")
	output := fs.String("o", "config.yml", "output config file")
	fs.Parse(args)

	if *input == "" {
		fmt.Fprintln(os.Stderr, "usage: specter gen -i openapi.yml [-o config.yml]")
		os.Exit(1)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(*input)
	if err != nil {
		log.Fatalf("failed to load OpenAPI spec: %v", err)
	}

	cfg := &config.Config{}

	paths := make([]string, 0, doc.Paths.Len())
	for path := range doc.Paths.Map() {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := doc.Paths.Value(path)
		ginPath := convertPath(path)

		// Sort methods for deterministic output
		ops := pathItem.Operations()
		methods := make([]string, 0, len(ops))
		for method := range ops {
			methods = append(methods, method)
		}
		sort.Strings(methods)

		for _, method := range methods {
			op := ops[method]
			if op.Responses == nil {
				continue
			}

			status, resp := firstSuccessResponse(op)
			if resp == nil {
				continue
			}

			route := config.Route{
				Path:   ginPath,
				Method: strings.ToUpper(method),
				Status: status,
			}

			for _, mediaType := range resp.Content {
				if ex := pickExample(mediaType); ex != nil {
					route.Response = ex
					break
				}
				if mediaType.Schema != nil && mediaType.Schema.Value != nil {
					route.Response = schemaToExample(mediaType.Schema.Value)
					break
				}
			}

			cfg.Routes = append(cfg.Routes, route)
		}
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(*output, out, 0o644); err != nil {
		log.Fatalf("failed to write %s: %v", *output, err)
	}

	fmt.Printf("generated %s (%d routes)\n", *output, len(cfg.Routes))
}

var paramRe = regexp.MustCompile(`\{([^}]+)\}`)

func convertPath(path string) string {
	return paramRe.ReplaceAllString(path, ":$1")
}

func firstSuccessResponse(op *openapi3.Operation) (int, *openapi3.Response) {
	for _, code := range []string{"200", "201", "202", "204"} {
		if ref := op.Responses.Value(code); ref != nil && ref.Value != nil {
			n, _ := strconv.Atoi(code)
			return n, ref.Value
		}
	}
	for code, ref := range op.Responses.Map() {
		if strings.HasPrefix(code, "2") && ref != nil && ref.Value != nil {
			n, _ := strconv.Atoi(code)
			return n, ref.Value
		}
	}
	return 0, nil
}

func pickExample(mt *openapi3.MediaType) any {
	if mt.Example != nil {
		return mt.Example
	}
	for _, ex := range mt.Examples {
		if ex.Value != nil {
			return ex.Value.Value
		}
	}
	return nil
}

func schemaToExample(schema *openapi3.Schema) any {
	if schema.Example != nil {
		return schema.Example
	}
	switch {
	case schema.Type.Is("string"):
		if len(schema.Enum) > 0 {
			return schema.Enum[0]
		}
		return "string"
	case schema.Type.Is("integer"):
		return 0
	case schema.Type.Is("number"):
		return 0.0
	case schema.Type.Is("boolean"):
		return true
	case schema.Type.Is("array"):
		if schema.Items != nil && schema.Items.Value != nil {
			return []any{schemaToExample(schema.Items.Value)}
		}
		return []any{}
	case schema.Type.Is("object"):
		obj := map[string]any{}
		props := make([]string, 0, len(schema.Properties))
		for name := range schema.Properties {
			props = append(props, name)
		}
		sort.Strings(props)
		for _, name := range props {
			ref := schema.Properties[name]
			if ref.Value != nil {
				obj[name] = schemaToExample(ref.Value)
			}
		}
		return obj
	}
	return nil
}
