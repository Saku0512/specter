# specter

Lightweight mock API server. Define endpoints in YAML, run instantly.

- Zero-config hot reload — edit `config.yml` and changes apply immediately, no restart needed
- Supports GET, POST, PUT, DELETE, PATCH and any other HTTP method
- Path parameters, custom status codes, arbitrary JSON responses

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/Saku0512/specter/main/install.sh | bash
```

## Usage

```sh
specter -c config.yml -p 8080
```

| Flag | Default | Description |
|------|---------|-------------|
| `-c` | `config.yaml` | Path to config file |
| `-p` | `8080` | Port to listen on |
| `--host` | all interfaces | Host to listen on |
| `-v`, `--version` | — | Show version |
| `--verbose` | — | Log request headers and body |

Flags take precedence over environment variables.

| Environment variable | Equivalent flag |
|----------------------|-----------------|
| `SPECTER_CONFIG` | `-c` |
| `SPECTER_PORT` | `-p` |
| `SPECTER_HOST` | `--host` |
| `SPECTER_VERBOSE` | `--verbose` |

## Generate config from OpenAPI

```sh
specter gen -i openapi.yml -o config.yml
```

| Flag | Default | Description |
|------|---------|-------------|
| `-i` | — | Path to OpenAPI spec (YAML or JSON) |
| `-o` | `config.yml` | Output config file |

- Converts `{param}` path parameters to `:param`
- Uses `example` / `examples` fields if defined
- Falls back to schema-based dummy values when no example is present

## Config

```yaml
routes:
  - path: /users
    method: GET
    status: 200
    response:
      - id: 1
        name: Alice
      - id: 2
        name: Bob

  - path: /users/:id
    method: GET
    status: 200
    response:
      id: 1
      name: Alice

  - path: /users
    method: POST
    status: 201
    response:
      message: created
```

Both `.yaml` and `.yml` extensions are supported. See [config.example.yml](config.example.yml) for a full example covering all features.

### Query Parameter Matching

Use `match` to return different responses based on query parameters. Falls back to the default `response` if no match.

```yaml
- path: /users
  method: GET
  match:
    - query:
        status: active
      response:
        - id: 1
          name: Alice
    - query:
        status: inactive
      status: 404
      response:
        error: not found
  response:
    - id: 1
    - id: 2
```

```sh
GET /users?status=active   → 200 [{ id: 1, name: Alice }]
GET /users?status=inactive → 404 { error: not found }
GET /users                 → 200 [{ id: 1 }, { id: 2 }]
```

### Path Parameters in Response

Use `:paramName` in response values to embed path parameters. Numeric values are automatically converted to numbers.

```yaml
- path: /users/:id
  method: GET
  response:
    id: ":id"       # /users/42 → { id: 42 }
    name: Alice
```

### Multiple Responses

Use `responses` to return different responses per request. Control the behavior with `mode`.

| mode | behavior |
|------|----------|
| `sequential` (default) | Returns responses in order, loops when exhausted |
| `random` | Picks a response randomly each time |

```yaml
# Retry simulation: fails first, succeeds on retry
- path: /unstable
  method: GET
  mode: sequential
  responses:
    - status: 500
      response: { error: internal }
    - status: 200
      response: { ok: true }

# Random failure simulation
- path: /flaky
  method: GET
  mode: random
  responses:
    - status: 200
      response: { ok: true }
    - status: 503
      response: { error: unavailable }
```

### CORS

Set `cors: true` to enable CORS headers for all routes. Preflight (`OPTIONS`) requests are handled automatically.

```yaml
cors: true

routes:
  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice
```

### Custom Response Headers

Add `headers` to set arbitrary response headers.

```yaml
routes:
  - path: /token
    method: POST
    headers:
      X-Auth-Token: dummy-token
      X-Request-Id: abc123
    response: { ok: true }
```

### Response Delay

Add `delay` (milliseconds) to simulate slow responses.

```yaml
routes:
  - path: /slow
    method: GET
    delay: 1000
    response:
      message: finally
```

### Verbose Logging

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

### Hot Reload

specter watches the config file and reloads automatically on save. No restart required.

```
[GIN] ...  👻 config reloaded
```

## License

MIT
