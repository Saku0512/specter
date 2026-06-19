# JSON Schema

Specter publishes a JSON Schema for config files at:

```text
https://specter.dev/schemas/config.schema.json
```

The canonical copy lives in this repository at:

```text
schemas/config.schema.json
```

The schema uses JSON Schema draft-07 and is intended for editor completion, inline validation, docs generation, and external tooling. It covers top-level config, routes, matchers, responses, stores, scenarios, includes, OpenAPI options, latency profiles, and fault profiles.

## Versioning

The current schema carries these metadata fields:

```json
{
  "x-specter-schema-version": "1",
  "x-specter-config-compatibility": "v0"
}
```

`schemas/config.schema.json` is the stable URL for the latest compatible Specter config schema. Breaking config-schema changes should increment `x-specter-schema-version`.

## VS Code

The Specter Mock Server VS Code extension automatically associates this schema with:

- `specter.yaml`
- `specter.yml`
- `config.yaml`
- `config.yml`

Without the extension, add this to VS Code settings if you use the Red Hat YAML extension:

```json
{
  "yaml.schemas": {
    "https://specter.dev/schemas/config.schema.json": [
      "specter.yaml",
      "specter.yml",
      "config.yaml",
      "config.yml"
    ]
  }
}
```

## Per-file Directive

For yaml-language-server compatible editors, add this comment to the top of a config file:

```yaml
# yaml-language-server: $schema=https://specter.dev/schemas/config.schema.json
routes:
  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice
```

## Local Development

Run the schema checks from the VS Code extension directory:

```sh
cd vscode-extension
npm run test:schema
```

The check verifies that:

- `schemas/config.schema.json` is valid JSON with expected metadata.
- The VS Code extension schema matches the canonical schema.
- The site-published schema copy matches the canonical schema.

Go tests also compare schema properties with the YAML tags in the config parser:

```sh
go test ./config
```
