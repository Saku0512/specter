package examples

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Example struct {
	Name        string
	Description string
	Files       map[string]string
}

const authConfig = `# specter auth example

scenarios:
  logged-out:
    state: ""
  logged-in:
    state: logged_in
    vars:
      role: admin

routes:
  - path: /login
    method: POST
    match:
      - body:
          email: "^admin@example.com$"
          password: "^password$"
        set_state: logged_in
        set_vars:
          role: admin
        response:
          access_token: dev-token
          token_type: bearer
    status: 401
    response:
      error: invalid_credentials

  - path: /me
    method: GET
    state: logged_in
    response:
      id: "1"
      email: admin@example.com
      role: "{{ .vars.role }}"

  - path: /me
    method: GET
    status: 401
    response:
      error: unauthorized
`

const crudConfig = `# specter CRUD example

scenarios:
  seeded:
    stores:
      todos:
        - id: "1"
          title: Write docs
          done: false
        - id: "2"
          title: Ship feature
          done: true

routes:
  - path: /todos
    method: POST
    store_push: todos

  - path: /todos
    method: GET
    store_list: todos

  - path: /todos/:id
    method: GET
    store_get: todos
    store_key: id

  - path: /todos/:id
    method: PATCH
    store_patch: todos
    store_key: id

  - path: /todos/:id
    method: DELETE
    store_delete: todos
    store_key: id
`

const paginationConfig = `# specter pagination example

stores:
  products:
    seed:
      - id: "1"
        name: Keyboard
        category: hardware
        price: 120
      - id: "2"
        name: Mouse
        category: hardware
        price: 60
      - id: "3"
        name: Notebook
        category: stationery
        price: 8

routes:
  - path: /products
    method: GET
    store_list: products
`

const openAPIConfig = `# specter OpenAPI example

openapi: ./openapi.yml
openapi_strict: true
openapi_strict_response: true

routes:
  - path: /pets
    method: GET
    response:
      - id: 1
        name: Fido

  - path: /pets
    method: POST
    status: 201
    response:
      id: 2
      name: "{{ .body.name }}"
`

const openAPISpec = `openapi: 3.0.0
info:
  title: Specter Pets API
  version: 1.0.0
paths:
  /pets:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  required: [id, name]
                  properties:
                    id:
                      type: integer
                    name:
                      type: string
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema:
                type: object
                required: [id, name]
                properties:
                  id:
                    type: integer
                  name:
                    type: string
`

const webhooksConfig = `# specter webhooks example

routes:
  - path: /orders
    method: POST
    status: 202
    response:
      id: ord_123
      status: accepted
    webhook:
      url: http://localhost:9999/hooks/orders
      method: POST
      delay: 250
      headers:
        X-Specter-Event: order.accepted
      body:
        order_id: ord_123
        status: accepted
`

const sseConfig = `# specter SSE example

routes:
  - path: /events
    method: GET
    stream: true
    stream_repeat: true
    events:
      - event: ready
        id: "1"
        data:
          status: connected
      - event: heartbeat
        id: "2"
        delay: 1000
        data:
          ok: true
`

const graphqlConfig = `# specter GraphQL example

routes:
  - path: /graphql
    method: POST
    match:
      - graphql:
          operation: "^GetUser$"
          variables:
            id: "^1$"
        response:
          data:
            user:
              id: "1"
              name: Alice
      - graphql:
          operation: "^ListUsers$"
        response:
          data:
            users:
              - id: "1"
                name: Alice
              - id: "2"
                name: Bob
    response:
      errors:
        - message: unsupported operation
`

const errorsConfig = `# specter common error states example

routes:
  - path: /unstable
    method: GET
    error_rate: 0.25
    error_status: 503
    response:
      ok: true

  - path: /limited
    method: GET
    rate_limit: 2
    rate_reset: 60
    response:
      ok: true

  - path: /missing
    method: GET
    status: 404
    response:
      error: not_found

  - path: /bad-request
    method: POST
    status: 400
    response:
      error: invalid_request
`

var inventory = []Example{
	{Name: "auth", Description: "Login, logout, state, vars, and unauthorized responses", Files: map[string]string{"config.yml": authConfig}},
	{Name: "crud", Description: "In-memory CRUD routes backed by a seeded store", Files: map[string]string{"config.yml": crudConfig}},
	{Name: "pagination", Description: "Store-backed list endpoint with filtering, sorting, and pagination query params", Files: map[string]string{"config.yml": paginationConfig}},
	{Name: "openapi", Description: "Request and response validation using a generated OpenAPI spec", Files: map[string]string{"config.yml": openAPIConfig, "openapi.yml": openAPISpec}},
	{Name: "webhooks", Description: "Async callback fired after a mock response", Files: map[string]string{"config.yml": webhooksConfig}},
	{Name: "sse", Description: "Server-Sent Events stream with repeatable events", Files: map[string]string{"config.yml": sseConfig}},
	{Name: "graphql", Description: "GraphQL operationName and variable matching", Files: map[string]string{"config.yml": graphqlConfig}},
	{Name: "errors", Description: "Common 400, 404, 429, and flaky 503-style responses", Files: map[string]string{"config.yml": errorsConfig}},
}

func Run(args []string) {
	if len(args) == 0 || args[0] == "list" || args[0] == "--list" || args[0] == "-l" {
		printList()
		return
	}

	name := args[0]
	fs := flag.NewFlagSet("examples "+name, flag.ExitOnError)
	output := fs.String("o", "config.yml", "output config file")
	force := fs.Bool("f", false, "overwrite existing files")
	fs.Parse(args[1:])

	written, err := writeExample(name, *output, *force)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("created %s example:\n", normalizedName(name))
	for _, path := range written {
		fmt.Println(" ", path)
	}
	fmt.Println("run: specter doctor -c", *output)
}

func printList() {
	fmt.Println("available examples:")
	for _, example := range sortedInventory() {
		fmt.Printf("  %-12s %s\n", example.Name, example.Description)
	}
	fmt.Println("\ncreate one with: specter examples <name> [-o config.yml] [-f]")
}

func sortedInventory() []Example {
	examples := append([]Example(nil), inventory...)
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Name < examples[j].Name
	})
	return examples
}

func exampleNames() []string {
	examples := sortedInventory()
	names := make([]string, 0, len(examples))
	for _, example := range examples {
		names = append(names, example.Name)
	}
	return names
}

func exampleFor(name string) (Example, bool) {
	normalized := normalizedName(name)
	for _, example := range inventory {
		if example.Name == normalized {
			return example, true
		}
	}
	return Example{}, false
}

func normalizedName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func writeExample(name, output string, force bool) ([]string, error) {
	example, ok := exampleFor(name)
	if !ok {
		return nil, fmt.Errorf("unknown example %q. Available examples: %s", name, strings.Join(exampleNames(), ", "))
	}

	targets := targetFiles(example, output)
	paths := make([]string, 0, len(targets))
	for path := range targets {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	if !force {
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return nil, fmt.Errorf("%s already exists. Use -f to overwrite", path)
			} else if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	for _, path := range paths {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(targets[path]), 0o644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", path, err)
		}
	}
	return paths, nil
}

func targetFiles(example Example, output string) map[string]string {
	dir := filepath.Dir(output)
	files := make(map[string]string, len(example.Files))
	for name, content := range example.Files {
		path := filepath.Join(dir, name)
		if name == "config.yml" {
			path = output
		}
		files[path] = content
	}
	return files
}
