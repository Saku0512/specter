# Specter VS Code Extension

VS Code support for Specter configuration files.

## Features

- Associates Specter config schemas with `specter.yaml`, `specter.yml`, `config.yaml`, and `config.yml`.
- Provides YAML completion for routes, methods, matchers, stores, vars, latency profiles, fault profiles, assertions-related response fields, and OpenAPI settings.
- Shows inline schema validation errors for invalid field names, enum values, status codes, and numeric ranges.

## Requirements

This extension depends on the Red Hat YAML extension because Specter config files are YAML.

## Development

Run these commands from this directory:

```sh
npm run check-types
npm run lint
npm run test:schema
npm run package
```

Press F5 in VS Code to launch an Extension Development Host.

## Release Notes

### 0.0.1

Initial schema-based Specter config support.
