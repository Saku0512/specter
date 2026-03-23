# Config Reference

Both `.yaml` and `.yml` extensions are supported. See [config.example.yml](../config.example.yml) for a full working example.

## Basic structure

```yaml
cors: true               # optional
proxy: http://api.example.com  # optional
openapi: ./openapi.yaml  # optional — enables request validation

routes:
  - path: /users
    method: GET
    status: 200
    response:
      - id: 1
        name: Alice
```

## Route fields

| Field | Type | Description |
|---|---|---|
| `path` | string | URL path, supports `:param` syntax |
| `method` | string | HTTP method (GET, POST, PUT, PATCH, DELETE, …) |
| `status` | int | Response status code (default: 200) |
| `response` | any | Response body (JSON object, array, or string) |
| `headers` | map | Custom response headers |
| `content_type` | string | Response Content-Type (default: `application/json`) |
| `delay` | int | Response delay in milliseconds |
| `on_call` | int | Only match on this call number (1-based); use on multiple routes with same path for retry simulation |
| `match` | list | Conditional responses by query/body/headers/body_path |
| `mode` | string | `sequential` (default) or `random` |
| `responses` | list | Multiple responses for cycling |
| `rate_limit` | int | Max requests before returning 429 |
| `rate_reset` | int | Seconds until rate limit counter resets |
| `state` | string | Only match when server is in this state |
| `set_state` | string | Transition server to this state after responding |
| `vars` | map | Only match when all specified vars equal the given values |
| `set_vars` | map | Set these vars after responding |
| `webhook` | object | Outgoing HTTP callback fired after responding |
| `file` | string | Path to a `.json`, `.yaml`, `.yml`, or text file to serve as the response body |
| `script` | string | Go template producing the response body (takes priority over `file` and `response`) |

---

## Query Parameter Matching

Use `match` to return different responses based on query parameters. Values are treated as **Go regular expressions** — use `^value$` for exact match. Falls back to the default `response` if no match.

```yaml
- path: /users
  method: GET
  match:
    - query:
        status: "^active$"
      response:
        - id: 1
          name: Alice
    - query:
        status: "^inactive$"
      status: 404
      response:
        error: not found
    - query:
        sort: "^(asc|desc)$"
      response:
        - id: 2
  response:
    - id: 1
    - id: 2
```

```
GET /users?status=active      → 200 [{ id: 1, name: Alice }]
GET /users?status=inactive    → 404 { error: not found }
GET /users?sort=asc           → 200 [{ id: 2 }]
GET /users                    → 200 [{ id: 1 }, { id: 2 }]
```

## Request Header Matching

Use `headers` in `match` to branch on request headers. Header names are case-insensitive; values are **Go regular expressions**.

```yaml
- path: /api/data
  method: GET
  match:
    - headers:
        Authorization: "^Bearer .+"
      response: { data: secret }
    - headers:
        X-Role: "^admin$"
      response: { data: admin-only }
  status: 401
  response: { error: unauthorized }
```

```
GET /api/data  Authorization: Bearer secret-token  → 200 { data: secret }
GET /api/data  X-Role: admin                       → 200 { data: admin-only }
GET /api/data                                      → 401 { error: unauthorized }
```

`headers`, `query`, `body`, and `body_path` can be combined in a single `match` entry (all conditions must match).

`match` entries also support `set_state` and `set_vars` to transition state after a specific match fires. Match-level values take priority over route-level values.

```yaml
- path: /login
  method: POST
  match:
    - body:
        user: alice
      set_state: logged_in      # only set when this match fires
      set_vars:
        role: admin
      status: 200
      response: { token: abc }
    - body:
        user: bob
      set_state: logged_in
      set_vars:
        role: viewer
      status: 200
      response: { token: xyz }
  status: 401
  response: { error: unauthorized }
```

## JSONPath / Regex Body Matching

Use `body_path` in `match` to match nested fields using dot-notation paths and regular expression patterns. All conditions use AND logic.

```yaml
- path: /orders
  method: POST
  match:
    - body_path:
        status: "^(pending|processing)$"
        user.role: "^admin$"
      status: 200
      response: { ok: true }
    - body_path:
        status: "^cancelled$"
      status: 422
      response: { error: order cancelled }
  response: { error: no match }
```

```
POST /orders {"status":"pending","user":{"role":"admin"}}   → 200 { ok: true }
POST /orders {"status":"cancelled"}                         → 422 { error: order cancelled }
POST /orders {"status":"other"}                             → 200 { error: no match }
```

- Paths use dot notation: `user.role` traverses `{ "user": { "role": "..." } }`
- Values are Go regular expressions (e.g. `^admin$`, `^(a|b)$`, `\d+`)
- A plain string like `admin` is treated as a regex and matches any string containing `admin`; use `^admin$` for exact match
- `body_path` can be combined with `query`, `body`, and `headers` in the same `match` entry

## Request Body Matching

Use `body` in `match` to branch on the request body. `query` and `body` can be combined (AND condition).

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

```
POST /users {"role":"admin"}  → 201 { id: 1, role: admin }
POST /users {"role":"guest"}  → 403 { error: forbidden }
POST /users {"role":"user"}   → 200 { id: 2 }
```

## Path Parameters in Response

Use `:paramName` in response values to embed path parameters. Numeric values are automatically converted to numbers.

```yaml
- path: /users/:id
  method: GET
  response:
    id: ":id"       # /users/42 → { id: 42 }
    name: Alice
```

## Call-number Matching (on_call)

Use `on_call` to match only on a specific call number (1-based). Useful for simulating retry scenarios without complex state management.

### Route-level on_call

Multiple routes with the same method+path are evaluated in order. A route with `on_call: N` only matches when it is the Nth call to that endpoint.

```yaml
routes:
  # Only matches on the 2nd call
  - path: /retry
    method: GET
    on_call: 2
    status: 200
    response: { ok: true }

  # Fallback for all other calls
  - path: /retry
    method: GET
    status: 503
    response: { error: unavailable }
```

```
GET /retry  (call 1) → 503 { error: unavailable }
GET /retry  (call 2) → 200 { ok: true }
GET /retry  (call 3) → 503 { error: unavailable }
```

### on_call inside responses

Set `on_call` on individual `responses[]` entries to pin them to a specific call number. Entries without `on_call` form the fallback pool for sequential/random cycling.

```yaml
- path: /items
  method: GET
  responses:
    - on_call: 1
      status: 503
      response: { error: first call fails }
    - on_call: 3
      status: 201
      response: { special: true }
    - status: 200
      response: { normal: true }
```

```
GET /items  (call 1) → 503  (on_call: 1 wins)
GET /items  (call 2) → 200  (fallback pool: normal)
GET /items  (call 3) → 201  (on_call: 3 wins)
GET /items  (call 4) → 200  (fallback pool: normal)
```

## Multiple Responses

Use `responses` to return different responses per request. Control the behavior with `mode`.

| mode | behavior |
|---|---|
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
```

## Response Script

Use `script` to generate a response body using a Go template. The template has access to the full request context, including `.body`, `.query`, `.params`, `.headers`, `.method`, and `.path`. If the output is valid JSON it is decoded automatically; otherwise the raw string is returned.

`script` takes priority over `file` and `response`, and can be used on routes, `responses[]` entries, and `match` entries.

```yaml
- path: /greet
  method: POST
  script: |
    {"msg": "Hello, {{ .body.name | upper }}!", "path": "{{ .path }}"}

- path: /users/:id
  method: GET
  script: |
    {"id": "{{ .params.id }}", "via": "{{ .method }}"}
```

### Template helpers

| Function | Example | Output |
|---|---|---|
| `upper` | `{{ upper "hello" }}` | `HELLO` |
| `lower` | `{{ lower "WORLD" }}` | `world` |
| `trim` | `{{ trim "  hi  " }}` | `hi` |
| `default` | `{{ default "anon" .body.name }}` | `anon` if name is empty |
| `now` | `{{ now }}` | `2024-03-15T10:30:00Z` |
| `add` | `{{ add 1 2 }}` | `3` |
| `sub` | `{{ sub 5 2 }}` | `3` |
| `fake` | `{{ fake "uuid" }}` | random UUID |

All [Faker](#faker) types are also available via `{{ fake "type" }}`.

### Dynamic status code from script

Include `"_status"` in the script's JSON output to set the HTTP status code dynamically. The key is removed from the response body.

```yaml
- path: /orders
  method: POST
  script: |
    {{- if eq .body.type "premium" -}}
    {"_status": 201, "tier": "premium"}
    {{- else -}}
    {"_status": 400, "error": "invalid type"}
    {{- end -}}
```

```
POST /orders {"type":"premium"}  → 201 { "tier": "premium" }
POST /orders {"type":"free"}     → 400 { "error": "invalid type" }
```

`_status` works in route-level `script`, `match[].script`, and `responses[].script`.

### Using script in match and responses

```yaml
- path: /echo
  method: POST
  match:
    - body:
        action: echo
      script: '{"echoed": "{{ .body.text }}", "at": "{{ now }}"}'
  response: { error: unknown action }

- path: /items
  method: GET
  responses:
    - script: '{"page": 1, "count": {{ fake "int" }}}'
    - script: '{"page": 2, "count": {{ fake "int" }}}'
```

## Response Templates

Use `{{ .body.field }}`, `{{ .query.param }}`, `{{ .params.name }}`, `{{ .headers.X-My-Header }}`, `{{ .method }}`, and `{{ .path }}` to embed request data in responses. For more complex logic use `script` instead.

```yaml
- path: /users
  method: POST
  response:
    name: "{{ .body.name }}"
    role: "{{ .body.role }}"

- path: /users/:id
  method: GET
  response:
    msg: "user {{ .params.id }}"
```

Template values are always strings. For numeric path parameters, the `:paramName` syntax auto-converts to numbers.

### Faker

Use `{{ fake "type" }}` to generate random values on every request:

```yaml
- path: /users
  method: GET
  response:
    id: '{{ fake "uuid" }}'
    name: '{{ fake "name" }}'
    email: '{{ fake "email" }}'
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

## Response Content Type

By default responses are served as `application/json`. Set `content_type` to return plain text, HTML, or any other MIME type.

```yaml
- path: /health
  method: GET
  content_type: text/plain
  response: "ok"

- path: /page
  method: GET
  content_type: text/html
  response: "<h1>Hello</h1>"
```

`content_type` can also be set per entry in `match` and `responses`.

## Rate Limit Simulation

Add `rate_limit` to return `429 Too Many Requests` after N requests. Use `rate_reset` to reset the counter after a given number of seconds.

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

## Stateful Mocking

Use `state` and `set_state` to simulate stateful flows like authentication. Multiple routes with the same method+path are matched in order — the first whose `state` condition matches wins. Routes without `state` always match.

```yaml
routes:
  - path: /login
    method: POST
    set_state: logged_in
    response: { token: abc }

  - path: /profile
    method: GET
    state: logged_in
    response: { name: Alice }

  - path: /profile
    method: GET
    status: 401
    response: { error: unauthorized }

  - path: /logout
    method: POST
    state: logged_in
    set_state: ""
    response: { ok: true }
```

```
POST /login   → 200 { token: abc }        (state → logged_in)
GET  /profile → 200 { name: Alice }
POST /logout  → 200 { ok: true }          (state → "")
GET  /profile → 401 { error: unauthorized }
```

See [introspection.md](introspection.md) for the `/__specter/state` endpoint.

## Multi-variable State

For scenarios that need multiple independent variables, use `vars` / `set_vars` alongside (or instead of) the single `state` field.

```yaml
routes:
  - path: /login
    method: POST
    set_vars:
      logged_in: "true"
      role: "{{ .body.role }}"   # template supported
    response: { token: abc }

  - path: /admin
    method: GET
    vars:
      logged_in: "true"
      role: admin
    response: { secret: data }

  - path: /admin
    method: GET
    vars:
      logged_in: "true"
    status: 403
    response: { error: forbidden }

  - path: /admin
    method: GET
    status: 401
    response: { error: unauthorized }
```

```
POST /login {"role":"admin"}  → 200   (logged_in=true, role=admin)
GET  /admin                   → 200 { secret: data }

POST /login {"role":"user"}   → 200   (logged_in=true, role=user)
GET  /admin                   → 403 { error: forbidden }
```

`vars` conditions use AND logic (all specified keys must match). Multiple routes with the same path are evaluated in order — first match wins.

See [introspection.md](introspection.md) for the `/__specter/vars` endpoint to read/write vars from tests.

## Chaos / Fault Injection

Simulate real-world unreliability to test client-side error handling, retries, and timeouts.

### Random errors

Use `error_rate` (0.0–1.0) to return an error response for a random fraction of requests.

```yaml
- path: /api
  method: GET
  error_rate: 0.3        # 30% of requests return an error
  error_status: 503      # status code for injected errors (default: 503)
  response: { ok: true } # returned for the other 70%
```

### Random delay

Use `delay_min` / `delay_max` to jitter response latency. Both fields must be set together.

```yaml
- path: /slow
  method: GET
  delay_min: 100   # minimum delay in ms
  delay_max: 800   # maximum delay in ms
  response: { ok: true }
```

`delay_min` / `delay_max` take precedence over the fixed `delay` field. All four fields can be combined:

```yaml
- path: /flaky
  method: GET
  delay_min: 200
  delay_max: 1000
  error_rate: 0.2
  error_status: 500
  response: { data: ok }
```

| Field | Type | Description |
|---|---|---|
| `error_rate` | float | Probability of error, 0.0–1.0 |
| `error_status` | int | HTTP status for injected error (default: 503) |
| `delay_min` | int | Minimum random delay in ms |
| `delay_max` | int | Maximum random delay in ms |

## File Response

Use `file` to serve a response body from an external file instead of inlining it in the config. Useful for large or complex fixtures.

Supported formats:

| Extension | Behaviour |
|---|---|
| `.json` | Parsed and served as `application/json` |
| `.yaml` / `.yml` | Parsed and served as `application/json` |
| anything else | Served as raw text; set `content_type` explicitly if needed |

```yaml
- path: /users
  method: GET
  file: fixtures/users.json    # served as application/json

- path: /health
  method: GET
  content_type: text/plain
  file: fixtures/health.txt

- path: /config
  method: GET
  file: fixtures/config.yaml   # YAML parsed → served as JSON
```

`file` can also be set per entry in `responses` and `match`:

```yaml
- path: /items
  method: GET
  mode: sequential
  responses:
    - file: fixtures/items_v1.json
    - file: fixtures/items_v2.json

- path: /search
  method: GET
  match:
    - query:
        type: premium
      file: fixtures/premium.json
  file: fixtures/default.json
```

File paths are resolved relative to the working directory where specter is started.

## Webhook / Callback

Use `webhook` to fire an outgoing HTTP request after specter responds. Useful for simulating event-driven systems (payment callbacks, notification services, async jobs).

```yaml
- path: /payments
  method: POST
  status: 202
  response: { status: processing }
  webhook:
    url: http://localhost:9000/payment-result
    method: POST                          # default: POST
    delay: 500                            # milliseconds before sending (default: 0)
    body:
      event: payment.completed
      amount: "{{ .body.amount }}"        # template from original request
    headers:
      X-Webhook-Secret: mysecret
```

The webhook is fired asynchronously — the original response is returned immediately without waiting.

### Webhook fields

| Field | Type | Description |
|---|---|---|
| `url` | string | Target URL (required; supports `{{ template }}`) |
| `method` | string | HTTP method (default: `POST`) |
| `body` | any | Request body; supports the same template syntax as responses |
| `headers` | map | Custom HTTP headers to include |
| `delay` | int | Milliseconds to wait before firing (default: 0) |

### Example: simulate async order fulfillment

```yaml
routes:
  - path: /orders
    method: POST
    status: 201
    response: { id: 42, status: pending }
    webhook:
      url: http://localhost:8080/fulfillment
      delay: 2000
      body:
        order_id: 42
        user: "{{ .body.user }}"
```

```
POST /orders {"user":"alice"}
→ 201 { id: 42, status: pending }

# 2 seconds later, specter sends:
POST http://localhost:8080/fulfillment {"order_id":42,"user":"alice"}
```

## OpenAPI Request Validation

Set `openapi` to a spec file path to enable non-blocking request validation. specter validates each incoming request against the spec and — if the request doesn't conform — adds a header and logs a warning, but **always serves the mock response regardless**.

```yaml
openapi: ./openapi.yaml

routes:
  - path: /items
    method: POST
    status: 201
    response: { id: 1 }
```

When a request fails validation:

```
X-Specter-Validation-Error: request body has an error: ... property "name" is missing
```

```
[specter] openapi validation: POST /items — request body has an error: ...
```

Supported spec formats: `.yaml`, `.yml`, `.json`. Routes not defined in the spec are silently skipped. Authentication checks are disabled by default (only schema validation runs).

This is useful for catching mismatches between your client and the API contract during development, without breaking your mock-based tests.

## Proxy Fallback

Set `proxy` to forward unmatched requests to a real API.

```yaml
proxy: http://api.example.com

routes:
  - path: /users
    method: GET
    response: [{ id: 1 }]   # served by specter
  # all other requests → forwarded to api.example.com
```

## CORS

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

## Custom Response Headers

```yaml
- path: /token
  method: POST
  headers:
    X-Auth-Token: dummy-token
    X-Request-Id: abc123
  response: { ok: true }
```

## Response Delay

```yaml
- path: /slow
  method: GET
  delay: 1000
  response:
    message: finally
```
