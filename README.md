# specter

Lightweight mock API server. Define endpoints in YAML, run instantly.

- Zero-config hot reload — edit `config.yml` and changes apply immediately, no restart needed
- Supports GET, POST, PUT, DELETE, PATCH and any other HTTP method
- Path parameters, custom status codes, arbitrary JSON responses

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

## Usage

```sh
specter -c config.yml -p 8080
```

| Flag | Default | Description |
|------|---------|-------------|
| `-c` | `config.yaml` | Path to config file |
| `-p` | `8080` | Port to listen on |
| `--host` | all interfaces | Host to listen on |
| `--cert` | — | TLS certificate file (enables HTTPS) |
| `--key` | — | TLS key file (enables HTTPS) |
| `-v`, `--version` | — | Show version |
| `--verbose` | — | Log request headers and body |

Flags take precedence over environment variables.

| Environment variable | Equivalent flag |
|----------------------|-----------------|
| `SPECTER_CONFIG` | `-c` |
| `SPECTER_PORT` | `-p` |
| `SPECTER_HOST` | `--host` |
| `SPECTER_CERT` | `--cert` |
| `SPECTER_KEY` | `--key` |
| `SPECTER_VERBOSE` | `--verbose` |

## Quick start

```sh
specter init          # generate config.yml in current directory
specter -c config.yml # start the server
```

`specter init -f` to overwrite an existing file.

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

### Request Body Matching

Use `body` in `match` to return different responses based on the request body. Can be combined with `query`.

```yaml
- path: /users
  method: POST
  match:
    - body:
        role: admin
      status: 201
      response: { id: 1, role: admin }
    - body:
        role: guest
      status: 403
      response: { error: forbidden }
  response: { id: 2 }
```

```sh
POST /users {"role":"admin"}  → 201 { id: 1, role: admin }
POST /users {"role":"guest"}  → 403 { error: forbidden }
POST /users {"role":"user"}   → 200 { id: 2 }
```

`query` と `body` は同時に指定できます（AND 条件）。

```yaml
match:
  - query:
      version: v2
    body:
      role: admin
    response: { ok: true }
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

### Proxy Fallback

Set `proxy` to forward unmatched requests to a real API.

```yaml
proxy: http://api.example.com

routes:
  - path: /users
    method: GET
    response: [{ id: 1 }]   # served by specter
  # all other requests → forwarded to api.example.com
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

### Response Templates

Use `{{ .body.field }}`, `{{ .query.param }}`, and `{{ .params.name }}` in response values to embed data from the request.

```yaml
- path: /users
  method: POST
  response:
    id: 1
    name: "{{ .body.name }}"
    role: "{{ .body.role }}"

- path: /search
  method: GET
  response:
    query: "{{ .query.q }}"
    results: []

- path: /users/:id
  method: GET
  response:
    msg: "user {{ .params.id }}"
```

```sh
POST /users {"name":"Alice","role":"admin"}  → { id: 1, name: "Alice", role: "admin" }
GET  /search?q=hello                         → { query: "hello", results: [] }
GET  /users/42                               → { msg: "user 42" }
```

Template values are always strings. For numeric path parameters, the existing `:paramName` syntax auto-converts to numbers.

### Response Content Type

By default, responses are served as `application/json`. Set `content_type` to return plain text, HTML, or any other MIME type.

```yaml
routes:
  - path: /health
    method: GET
    content_type: text/plain
    response: "ok"

  - path: /page
    method: GET
    content_type: text/html
    response: "<h1>Hello</h1>"

  - path: /data
    method: GET
    content_type: application/xml
    response: "<user><id>1</id></user>"
```

`content_type` can also be set per entry in `match` and `responses`, overriding the route-level value.

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

### HTTPS

Pass `--cert` and `--key` to enable TLS.

```sh
specter -c config.yml --cert cert.pem --key key.pem
# 👻 Specter running on :8080 (TLS)
```

For local development, you can generate a self-signed certificate with:

```sh
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj '/CN=localhost'
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
