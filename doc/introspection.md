# Introspection API

specter exposes built-in endpoints under `/__specter/` for debugging and test automation. These routes are never recorded in the request history.

## Request History

Records incoming requests in memory (up to 200 entries, oldest dropped when full).

```sh
GET    /__specter/requests        # list all recorded requests
GET    /__specter/requests/:index # get request by index (0-based)
DELETE /__specter/requests        # clear history
```

Example response for `GET /__specter/requests`:

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

## Request Verification

`POST /__specter/requests/assert` checks that recorded requests match the given criteria. Useful in CI/E2E tests to verify your app made the expected API calls.

### Request body

| Field | Type | Description |
|---|---|---|
| `method` | string | HTTP method to match (case-insensitive) |
| `path` | string | Exact path to match |
| `query` | object | Query params that must be present (subset match) |
| `body` | object | JSON fields that must be present in the request body (subset match) |
| `count` | int | Exact number of matching requests expected. Omit to require at least 1. |

### Examples

```sh
# Assert /users was called at least once
curl -X POST http://localhost:8080/__specter/requests/assert \
  -H 'Content-Type: application/json' \
  -d '{"path":"/users"}'

# Assert POST /users was called with name=Alice exactly once
curl -X POST http://localhost:8080/__specter/requests/assert \
  -d '{"method":"POST","path":"/users","body":{"name":"Alice"},"count":1}'

# Assert /admin was never called
curl -X POST http://localhost:8080/__specter/requests/assert \
  -d '{"path":"/admin","count":0}'
```

### Responses

| Status | Meaning |
|---|---|
| `200` | Assertion passed — `{ "ok": true, "matched": N }` |
| `422` | Assertion failed — `{ "ok": false, "matched": N, "error": "..." }` |

## Dynamic Routes

Add, list, and remove routes at runtime without editing the config file or restarting. Useful for per-test scenario setup in CI/E2E.

```sh
GET    /__specter/routes        # list all routes (config + dynamic)
POST   /__specter/routes        # add a route → returns { "id": "<uuid>" }
DELETE /__specter/routes        # remove all dynamic routes
DELETE /__specter/routes/:id    # remove one dynamic route by ID
```

`POST /__specter/routes` accepts the same fields as a config route:

```sh
curl -X POST http://localhost:8080/__specter/routes \
  -H 'Content-Type: application/json' \
  -d '{
    "path": "/feature-flag",
    "method": "GET",
    "status": 200,
    "response": { "enabled": true }
  }'
# → { "id": "550e8400-e29b-41d4-a716-446655440000" }
```

Dynamic routes are merged with config routes and processed in order. Config routes are listed with `"source": "config"`; dynamic routes with `"source": "dynamic"` and an `"id"` field.

Dynamic routes persist across hot reloads but are cleared when the server restarts.

## State

Read or override the current [stateful mocking](config.md#stateful-mocking) state.

```sh
GET /__specter/state              # { "state": "logged_in" }
PUT /__specter/state              # set state
```

```sh
# Reset state before each test
curl -X PUT http://localhost:8080/__specter/state \
  -H 'Content-Type: application/json' \
  -d '{"state":""}'
```

The state persists across hot reloads but resets when the server restarts.

## Vars (multi-variable state store)

A key-value store for more complex stateful scenarios. See [config.md](config.md#multi-variable-state) for route-level `vars` / `set_vars` fields.

```sh
GET    /__specter/vars           # get all vars  → { "role": "admin", ... }
PUT    /__specter/vars           # set multiple  ← { "role": "admin", "tier": "gold" }
DELETE /__specter/vars           # clear all vars

GET    /__specter/vars/:key      # get one var   → { "key": "role", "value": "admin" }
PUT    /__specter/vars/:key      # set one var   ← { "value": "admin" }
DELETE /__specter/vars/:key      # delete one var
```

Vars persist across hot reloads but reset when the server restarts.

## Reset

Reset state, vars, and request history in a single call. Useful for test setup/teardown.

```sh
POST /__specter/reset        # reset everything (state + vars + history)
```

With an optional `targets` array to reset selectively:

```sh
curl -X POST http://localhost:8080/__specter/reset \
  -H 'Content-Type: application/json' \
  -d '{"targets": ["state", "vars", "history"]}'
```

| Target | Resets |
|---|---|
| `state` | Server state (same as `PUT /__specter/state {"state":""}`) |
| `vars` | All vars (same as `DELETE /__specter/vars`) |
| `history` | Request history (same as `DELETE /__specter/requests`) |

Omit `targets` (or send `{}`) to reset all three at once.

Dynamic routes are **not** affected — use `DELETE /__specter/routes` to clear those separately.

```sh
# Reset before each test
curl -X POST http://localhost:8080/__specter/reset
# → { "ok": true }
```
