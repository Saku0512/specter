# Changelog

All notable changes to the Specter Mock Server VS Code extension will be documented in this file.

## [0.0.3] - 2026-06-20

- Publish the Specter config JSON Schema as a canonical repository artifact.
- Bundle the published schema into the VS Code extension for editor completion and validation.
- Document schema usage for VS Code and yaml-language-server compatible editors.
- Keep the extension schema, site schema, and canonical schema synchronized with validation checks.

## [0.0.2] - 2026-06-20

- Rename the extension to Specter Mock Server.
- Add Marketplace metadata and a 128x128 extension icon.
- Add GitHub Actions packaging that builds a patch-versioned VSIX artifact without publishing automatically.

## [0.0.1] - 2026-06-20

- Add schema association for `specter.yaml`, `specter.yml`, `config.yaml`, and `config.yml`.
- Add completion and schema validation for Specter routes, matchers, stores, scenarios, OpenAPI settings, latency profiles, and fault profiles.
- Add extension development, schema validation, packaging, and test scripts.
