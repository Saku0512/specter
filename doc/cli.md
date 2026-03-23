# CLI Reference

## Usage

```sh
specter -c config.yml -p 8080
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `-c` | `config.yaml` | Path to config file |
| `-p` | `8080` | Port to listen on |
| `--host` | all interfaces | Host to listen on |
| `--cert` | — | TLS certificate file (enables HTTPS) |
| `--key` | — | TLS key file (enables HTTPS) |
| `--ui-port` | `4444` | Port for the web UI (set to `0` to disable) |
| `--verbose` | — | Log request headers and body |
| `-v`, `--version` | — | Show version |
| `-h`, `--help` | — | Show help |

Flags take precedence over environment variables.

## Environment Variables

| Variable | Equivalent flag |
|---|---|
| `SPECTER_CONFIG` | `-c` |
| `SPECTER_PORT` | `-p` |
| `SPECTER_HOST` | `--host` |
| `SPECTER_CERT` | `--cert` |
| `SPECTER_KEY` | `--key` |
| `SPECTER_VERBOSE` | `--verbose` |
| `SPECTER_UI_PORT` | `--ui-port` |

## Web UI

specter ships with a minimal built-in dashboard, served on port `4444` by default. Open it in your browser while the server is running:

```
http://localhost:4444
```

The UI shows four tabs that auto-refresh every 2 seconds:

| Tab | What it shows |
|---|---|
| **Requests** | Recorded request history (newest first). Includes a Clear button. |
| **Routes** | All registered routes (config + dynamic), with source and metadata badges. |
| **State & Vars** | Current server state and all var values. |
| **Stores** | Contents of every in-memory CRUD store collection. |

To disable the UI, pass `--ui-port 0` or set `SPECTER_UI_PORT=0`.

## Subcommands

### `specter init`

Generate a starter `config.yml` in the current directory.

```sh
specter init          # create config.yml
specter init -f       # overwrite if it already exists
specter init -o my.yml
```

### `specter gen`

Generate a config file from an OpenAPI spec.

```sh
specter gen -i openapi.yml -o config.yml
```

| Flag | Default | Description |
|---|---|---|
| `-i` | — | Path to OpenAPI spec (YAML or JSON) |
| `-o` | `config.yml` | Output config file |

- Converts `{param}` path parameters to `:param`
- Uses `example` / `examples` fields if defined
- Falls back to schema-based dummy values when no example is present

### `specter validate`

Validate a config file and report errors.

```sh
specter validate -c config.yml
```

### `specter record`

Proxy a real API and automatically generate `config.yml` from the recorded responses.

```sh
specter record -t http://api.example.com -o config.yml
```

Send requests through the recorder (e.g. with curl or your app), then press Ctrl+C to save.

```sh
curl http://localhost:8080/users
curl http://localhost:8080/users/1
# ^C
# ✓ recorded 2 route(s) → config.yml
```

| Flag | Default | Description |
|---|---|---|
| `-t` | — | Target URL to proxy to (required) |
| `-o` | `config.yml` | Output config file |
| `-p` | `8080` | Port to listen on |
| `-f` | — | Overwrite output file if it exists |

- Each (method, path) pair is recorded once (first response wins)
- JSON responses are stored as structured data
- Non-JSON responses include `content_type` automatically
- CORS preflight (`OPTIONS`) requests are forwarded but not recorded

## HTTPS

Pass `--cert` and `--key` to enable TLS.

```sh
specter -c config.yml --cert cert.pem --key key.pem
# 👻 Specter running on :8080 (TLS)
```

Generate a self-signed certificate for local development:

```sh
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj '/CN=localhost'
```

## Verbose Logging

Run with `--verbose` to log request headers and body for every request.

```sh
specter -c config.yml --verbose
```

```
→ POST /users
  Content-Type: application/json
  Authorization: Bearer token
  Body: {"name":"Alice"}
```

## Hot Reload

specter watches the config file and reloads automatically on save. No restart required.

```
[GIN] ...  👻 config reloaded
```
