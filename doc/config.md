# Config Reference

Both `.yaml` and `.yml` extensions are supported. See [config.example.yml](../config.example.yml) for a full working example.

## Basic structure

```yaml
cors: true               # optional
proxy: http://api.example.com  # optional

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
| `match` | list | Conditional responses by query/body/headers |
| `mode` | string | `sequential` (default) or `random` |
| `responses` | list | Multiple responses for cycling |
| `rate_limit` | int | Max requests before returning 429 |
| `rate_reset` | int | Seconds until rate limit counter resets |
| `state` | string | Only match when server is in this state |
| `set_state` | string | Transition server to this state after responding |
| `webhook` | object | Outgoing HTTP callback fired after responding |
| `file` | string | Path to a `.json`, `.yaml`, `.yml`, or text file to serve as the response body |

---

## Query Parameter Matching

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

```
GET /users?status=active   → 200 [{ id: 1, name: Alice }]
GET /users?status=inactive → 404 { error: not found }
GET /users                 → 200 [{ id: 1 }, { id: 2 }]
```

## Request Header Matching

Use `headers` in `match` to branch on request headers. Matching is case-insensitive on header names.

```yaml
- path: /api/data
  method: GET
  match:
    - headers:
        Authorization: Bearer secret-token
      response: { data: secret }
    - headers:
        X-Role: admin
      response: { data: admin-only }
  status: 401
  response: { error: unauthorized }
```

```
GET /api/data  Authorization: Bearer secret-token  → 200 { data: secret }
GET /api/data  X-Role: admin                       → 200 { data: admin-only }
GET /api/data                                      → 401 { error: unauthorized }
```

`headers`, `query`, and `body` can be combined in a single `match` entry (all conditions must match).

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

## Response Templates

Use `{{ .body.field }}`, `{{ .query.param }}`, and `{{ .params.name }}` to embed request data in responses.

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
