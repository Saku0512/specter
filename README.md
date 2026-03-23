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

## Record from a real API

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
|------|---------|-------------|
| `-t` | — | Target URL to proxy to (required) |
| `-o` | `config.yml` | Output config file |
| `-p` | `8080` | Port to listen on |
| `-f` | — | Overwrite output file if it exists |

- Each (method, path) pair is recorded once (first response wins)
- JSON responses are stored as structured data
- Non-JSON responses include `content_type` automatically
- CORS preflight (`OPTIONS`) requests are forwarded but not recorded

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

Use `{{ fake "type" }}` to generate random values on every request:

```yaml
- path: /users
  method: GET
  response:
    id: '{{ fake "uuid" }}'
    name: '{{ fake "name" }}'
    email: '{{ fake "email" }}'
    company: '{{ fake "company" }}'
```

| Type | Example output |
|---|---|
| `name` | `John Doe` |
| `first_name` | `John` |
| `last_name` | `Doe` |
| `email` | `john@example.com` |
| `uuid` | `550e8400-e29b-41d4-a716-446655440000` |
| `phone` | `555-123-4567` |
| `url` | `https://example.com` |
| `ip` | `192.168.1.1` |
| `username` | `johndoe42` |
| `password` | `Abc123xyz` |
| `word` | `cloud` |
| `sentence` | `The quick brown fox jumps.` |
| `paragraph` | `Lorem ipsum...` |
| `color` | `crimson` |
| `country` | `Japan` |
| `city` | `Tokyo` |
| `zip` | `100-0001` |
| `street` | `123 Main St` |
| `company` | `Acme Corp` |
| `job` | `Software Engineer` |
| `int` | `4821` |
| `float` | `73.42` |
| `bool` | `true` |
| `date` | `2024-03-15` |
| `datetime` | `2024-03-15T10:30:00Z` |

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

### Rate Limit Simulation

Add `rate_limit` to return `429 Too Many Requests` after N requests. Use `rate_reset` to automatically reset the counter after a given number of seconds.

```yaml
# Allow 5 requests, then always return 429
- path: /api
  method: GET
  rate_limit: 5
  response: { ok: true }

# Allow 10 requests per minute
- path: /api/windowed
  method: GET
  rate_limit: 10
  rate_reset: 60
  response: { ok: true }
```

When `rate_reset` is set, a `Retry-After` header is included in 429 responses.

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

### Stateful Mocking

Use `state` and `set_state` to simulate stateful flows like authentication.

```yaml
routes:
  # Login: always accessible, transitions to logged_in
  - path: /login
    method: POST
    set_state: logged_in
    response: { token: abc }

  # Profile: only accessible when logged_in, fallback to 401
  - path: /profile
    method: GET
    state: logged_in
    response: { name: Alice }

  - path: /profile
    method: GET
    status: 401
    response: { error: unauthorized }

  # Logout: only when logged_in, resets state
  - path: /logout
    method: POST
    state: logged_in
    set_state: ""
    response: { ok: true }
```

```sh
POST /login            → 200 { token: abc }      (state → logged_in)
GET  /profile          → 200 { name: Alice }
POST /logout           → 200 { ok: true }         (state → "")
GET  /profile          → 401 { error: unauthorized }
```

Multiple routes with the same method+path are matched in order — the first whose `state` condition matches wins. Routes without `state` always match.

**Built-in state endpoints:**

```sh
GET /__specter/state             # { "state": "logged_in" }
PUT /__specter/state {"state":""} # reset state (useful in test setup)
```

### Request History

specter records incoming requests in memory (up to 200 entries). Use the built-in endpoints to inspect or clear the history.

```sh
GET    /__specter/requests   # list recorded requests
DELETE /__specter/requests   # clear history
```

Example response:

```json
[
  {
    "time": "2024-01-01T00:00:00Z",
    "method": "POST",
    "path": "/users",
    "query": { "v": "2" },
    "headers": { "Content-Type": "application/json" },
    "body": "{\"name\":\"Alice\"}"
  }
]
```

`/__specter/*` routes are never recorded in the history.

### Hot Reload

specter watches the config file and reloads automatically on save. No restart required.

```
[GIN] ...  👻 config reloaded
```

## License

MIT
