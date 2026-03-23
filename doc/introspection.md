# Introspection API

specter exposes built-in endpoints under `/__specter/` for debugging and test automation. These routes are never recorded in the request history.

## Request History

Records incoming requests in memory (up to 200 entries, oldest dropped when full).

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

Useful in CI to assert that the app made expected API calls:

```sh
# After running your tests:
curl http://localhost:8080/__specter/requests | jq '.[0].path'
```

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
