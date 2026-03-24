# specter

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Saku0512/specter)
[![CI](https://github.com/Saku0512/specter/actions/workflows/test.yml/badge.svg)](https://github.com/Saku0512/specter/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Saku0512/specter)](https://goreportcard.com/report/github.com/Saku0512/specter)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Saku0512/specter)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/Saku0512/specter)](https://github.com/Saku0512/specter/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

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

## Quick start

```sh
specter init          # generate config.yml
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

- [Config reference](doc/config.md) — routes, matching, templates, faker, state, rate limiting, …
- [CLI reference](doc/cli.md) — flags, env vars, `init` / `gen` / `validate` / `record`
- [Introspection API](doc/introspection.md) — `/__specter/requests`, `/__specter/state`

See [config.example.yml](config.example.yml) for a full working example.

## License

MIT
