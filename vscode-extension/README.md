# Specter Config VS Code Extension

VS Code support for Specter configuration files.

## Features

- Associates Specter config schemas with `specter.yaml`, `specter.yml`, `config.yaml`, and `config.yml`.
- Provides YAML completion for routes, methods, matchers, stores, vars, latency profiles, fault profiles, assertions-related response fields, and OpenAPI settings.
- Shows inline schema validation errors for invalid field names, enum values, status codes, and numeric ranges.

## Requirements

This extension depends on the Red Hat YAML extension because Specter config files are YAML. VS Code installs extension dependencies automatically when the extension is installed from the Marketplace.

## Development

Run these commands from this directory:

```sh
npm run check-types
npm run lint
npm run test:schema
npm run package
npm run vsix
```

Press F5 in VS Code to launch an Extension Development Host.

## Publishing

This extension is ready to package with `vsce`:

```sh
npm run vsix
```

Publishing to Visual Studio Marketplace requires a Marketplace publisher named `saku0512` or updating the `publisher` field in `package.json` to match the publisher ID you create.

## Release Notes

### 0.0.1

Initial schema-based Specter config support.
