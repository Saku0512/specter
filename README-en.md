# specter

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Saku0512/specter)
[![CI](https://github.com/Saku0512/specter/actions/workflows/test.yml/badge.svg)](https://github.com/Saku0512/specter/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Saku0512/specter)](https://goreportcard.com/report/github.com/Saku0512/specter)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/Saku0512/specter/badge)](https://scorecard.dev/viewer/?uri=github.com/Saku0512/specter)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Saku0512/specter)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/Saku0512/specter)](https://github.com/Saku0512/specter/releases/latest)
[![VS Code Marketplace](https://img.shields.io/visual-studio-marketplace/v/Saku0512-sec.specter-mock-server?label=VS%20Code%20Extension)](https://marketplace.visualstudio.com/items?itemName=Saku0512-sec.specter-mock-server)
[![VS Code Installs](https://img.shields.io/visual-studio-marketplace/i/Saku0512-sec.specter-mock-server)](https://marketplace.visualstudio.com/items?itemName=Saku0512-sec.specter-mock-server)
![Downloads](https://img.shields.io/github/downloads/Saku0512/specter/total)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[日本語](README.md)

Lightweight mock API server. Define endpoints in YAML, run instantly.

- Hot reload — edit `config.yml` and changes apply immediately
- Response templates, faker, stateful mocking, rate limiting
- Single binary, no dependencies

## Install

**Docker**

```sh
docker run -v $(pwd)/config.yml:/config.yml ghcr.io/saku0512/specter -c /config.yml
```

**Homebrew (macOS / Linux)**

```sh
brew tap Saku0512/specter https://github.com/Saku0512/specter
brew install specter
```

**curl (macOS / Linux)**

```sh
curl -fsSL https://raw.githubusercontent.com/Saku0512/specter/main/install.sh | bash
```

**PowerShell (Windows)**

```powershell
irm https://raw.githubusercontent.com/Saku0512/specter/main/install.ps1 | iex
```

## Release Verification

Releases include `SHA256SUMS.txt` and an SPDX JSON SBOM. `install.sh` and `install.ps1` verify the binary checksum before installing it.

To verify manually:

```sh
VERSION=v1.0.1
ASSET=specter_linux_amd64
curl -LO "https://github.com/Saku0512/specter/releases/download/${VERSION}/${ASSET}"
curl -LO "https://github.com/Saku0512/specter/releases/download/${VERSION}/SHA256SUMS.txt"
grep "  ${ASSET}$" SHA256SUMS.txt | sha256sum --check -
gh attestation verify "${ASSET}" --repo Saku0512/specter
```

## Quick start

```sh
specter init          # generate config.yml
specter init --template crud  # generate a CRUD starter
specter examples      # list sample configs
specter doctor -c config.yml  # diagnose config, file references, and ports
specter -c config.yml # start the server
```

```yaml
routes:
  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice
```

## Documentation

- [Config reference](doc/config.md) — routes, matching, templates, faker, state, rate limiting, ...
- [Examples gallery](doc/examples.md) — auth, CRUD, pagination, GraphQL, webhooks, SSE, OpenAPI, and more
- [Comparison guide](doc/comparison.md) — differences from json-server, Prism, and WireMock
- [JSON Schema](doc/schema.md) — schema for editor completion and validation
- [VS Code Extension](https://marketplace.visualstudio.com/items?itemName=Saku0512-sec.specter-mock-server) — completion and inline validation for Specter config files
- [CLI reference](doc/cli.md) — flags, env vars, `init` / `examples` / `gen` / `validate` / `doctor` / `record`
- [Introspection API](doc/introspection.md) — `/__specter/requests`, `/__specter/state`

See [config.example.yml](config.example.yml) for a full working example.

## Development

Run the regular test suite with `make test`. To run the Go fuzz targets locally:

```sh
make fuzz
FUZZTIME=30s make fuzz
```

The fuzz targets exercise the config YAML parser and request matchers in memory, so they do not require network access.

## License

MIT
