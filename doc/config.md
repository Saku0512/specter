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
| `match` | list | Conditional responses by query/body |
| `mode` | string | `sequential` (default) or `random` |
| `responses` | list | Multiple responses for cycling |
| `rate_limit` | int | Max requests before returning 429 |
| `rate_reset` | int | Seconds until rate limit counter resets |
| `state` | string | Only match when server is in this state |
| `set_state` | string | Transition server to this state after responding |

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
