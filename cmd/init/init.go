package init_cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const basicTemplate = `# specter config
# docs: https://github.com/Saku0512/specter

routes:
  - path: /hello
    method: GET
    response:
      message: Hello, World!

  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice
      - id: 2
        name: Bob

  - path: /users/:id
    method: GET
    response:
      id: ":id"
      name: Alice

  - path: /users
    method: POST
    status: 201
    response:
      message: created
`

const crudTemplate = `# specter CRUD starter
# docs: https://github.com/Saku0512/specter

scenarios:
  seeded:
    stores:
      users:
        - id: "1"
          name: Alice
          role: admin
        - id: "2"
          name: Bob
          role: user

routes:
  - path: /users
    method: POST
    store_push: users

  - path: /users
    method: GET
    store_list: users

  - path: /users/:id
    method: GET
    store_get: users
    store_key: id

  - path: /users/:id
    method: PUT
    store_put: users
    store_key: id

  - path: /users/:id
    method: PATCH
    store_patch: users
    store_key: id

  - path: /users/:id
    method: DELETE
    store_delete: users
    store_key: id
`

const authTemplate = `# specter auth starter
# docs: https://github.com/Saku0512/specter

scenarios:
  logged-in-admin:
    state: logged_in
    vars:
      role: admin
    stores:
      users:
        - id: "1"
          name: Alice
          role: admin
  logged-out:
    state: ""
    vars: {}
    stores: {}

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
          access_token: test-token
          token_type: bearer
    status: 401
    response:
      error: invalid_credentials

  - path: /me
    method: GET
    state: logged_in
    response:
      id: "1"
      name: Alice
      role: admin

  - path: /me
    method: GET
    status: 401
    response:
      error: unauthorized

  - path: /logout
    method: POST
    state: logged_in
    set_state: ""
    set_vars:
      role: ""
    response:
      ok: true
`

const openAPITemplate = `# specter OpenAPI starter
# docs: https://github.com/Saku0512/specter

openapi: ./openapi.yml
openapi_strict: false
openapi_strict_response: false

routes:
  - path: /pets
    method: GET
    response:
      - id: 1
        name: Fido
        species: dog

  - path: /pets/:id
    method: GET
    response:
      id: ":id"
      name: Fido
      species: dog

  - path: /pets
    method: POST
    status: 201
    response:
      id: 2
      name: "{{ .body.name }}"
      species: "{{ .body.species }}"
`

var templates = map[string]string{
	"basic":   basicTemplate,
	"crud":    crudTemplate,
	"auth":    authTemplate,
	"openapi": openAPITemplate,
}

func templateNames() []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func templateFor(name string) (string, bool) {
	tmpl, ok := templates[strings.ToLower(strings.TrimSpace(name))]
	return tmpl, ok
}

func normalizedTemplateName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func writeConfig(output, templateName string, force bool) (string, error) {
	normalized := normalizedTemplateName(templateName)
	template, ok := templateFor(normalized)
	if !ok {
		return "", fmt.Errorf("unknown template %q. Available templates: %s", templateName, strings.Join(templateNames(), ", "))
	}

	if _, err := os.Stat(output); err == nil && !force {
		return "", fmt.Errorf("%s already exists. Use -f to overwrite", output)
	}

	if err := os.WriteFile(output, []byte(template), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", output, err)
	}
	return normalized, nil
}

func Run(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("o", "config.yml", "output file")
	force := fs.Bool("f", false, "overwrite if file already exists")
	templateName := fs.String("template", "basic", "starter template: basic, crud, auth, openapi")
	listTemplates := fs.Bool("list-templates", false, "list available starter templates")
	fs.Parse(args)

	if *listTemplates {
		fmt.Println("available templates:")
		for _, name := range templateNames() {
			fmt.Println(" ", name)
		}
		return
	}

	normalized, err := writeConfig(*output, *templateName, *force)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("created %s from %s template\n", *output, normalized)
	fmt.Println("run: specter -c", *output)
}
