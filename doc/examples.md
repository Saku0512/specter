# Examples Gallery

This gallery collects common Specter mocking patterns and the quickest way to try each one. Every example can be generated with `specter examples <name>` and then checked with `specter doctor -c config.yml`.

```sh
specter examples
specter examples auth -o config.yml -f
specter doctor -c config.yml
specter -c config.yml
```

## Gallery

| Example | Generate | Use when you need |
|---|---|---|
| [Auth](#auth) | `specter examples auth` | Login, protected endpoints, state, vars, and 401 responses |
| [CRUD](#crud) | `specter examples crud` | A REST-ish API backed by Specter's in-memory store |
| [Pagination](#pagination) | `specter examples pagination` | List endpoints with filtering, sorting, limit, and offset |
| [GraphQL](#graphql) | `specter examples graphql` | Branching by `operationName` and GraphQL variables |
| [Webhooks](#webhooks) | `specter examples webhooks` | Async callbacks after a mock response |
| [SSE](#sse) | `specter examples sse` | Server-Sent Events and repeated event streams |
| [OpenAPI](#openapi) | `specter examples openapi` | Request and response validation against an OpenAPI spec |
| [Polling / Long-Running Job](#polling--long-running-job) | copy from this page | A job endpoint that changes over repeated calls |
| [Error And Latency Scenarios](#error-and-latency-scenarios) | `specter examples errors` | 400/404/429 responses, flaky 503s, and slow APIs |

## Auth

Use state and vars to model a login flow. The first matching route wins, so the protected `/me` route can return either the logged-in profile or a fallback 401.

```yaml
scenarios:
  logged-out:
    state: ""
  logged-in:
    state: logged_in
    vars:
      role: admin

routes:
  - path: /login
    method: POST
    match:
      - body:
          email: "^admin@example.com$"
          password: "^password$"
        set_state: logged_in
        set_vars:
          role: admin
        response:
          access_token: dev-token
          token_type: bearer
    status: 401
    response:
      error: invalid_credentials

  - path: /me
    method: GET
    state: logged_in
    response:
      id: "1"
      email: admin@example.com
      role: "{{ .vars.role }}"

  - path: /me
    method: GET
    status: 401
    response:
      error: unauthorized
```

Try it:

```sh
specter examples auth -o config.yml -f
specter -c config.yml
curl -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"password"}'
curl http://localhost:8080/me
```

## CRUD

Wire REST endpoints directly to a named in-memory store. This is a fast way to unblock frontend list/detail/create/edit/delete screens before a backend exists.

```yaml
scenarios:
  seeded:
    stores:
      todos:
        - id: "1"
          title: Write docs
          done: false
        - id: "2"
          title: Ship feature
          done: true

routes:
  - path: /todos
    method: POST
    store_push: todos

  - path: /todos
    method: GET
    store_list: todos

  - path: /todos/:id
    method: GET
    store_get: todos
    store_key: id

  - path: /todos/:id
    method: PATCH
    store_patch: todos
    store_key: id

  - path: /todos/:id
    method: DELETE
    store_delete: todos
    store_key: id
```

Try it:

```sh
specter examples crud -o config.yml -f
specter -c config.yml
curl http://localhost:8080/todos
curl -X POST http://localhost:8080/todos \
  -H 'Content-Type: application/json' \
  -d '{"title":"Demo the flow","done":false}'
```

## Pagination

`store_list` supports query parameters for filtering, sorting, and pagination. Seed a store and point a list endpoint at it.

```yaml
stores:
  products:
    seed:
      - id: "1"
        name: Keyboard
        category: hardware
        price: 120
      - id: "2"
        name: Mouse
        category: hardware
        price: 60
      - id: "3"
        name: Notebook
        category: stationery
        price: 8

routes:
  - path: /products
    method: GET
    store_list: products
```

Try common list queries:

```sh
curl 'http://localhost:8080/products?category=hardware'
curl 'http://localhost:8080/products?_sort=price&_order=desc'
curl 'http://localhost:8080/products?_limit=10&_offset=20'
```

## GraphQL

Use `match.graphql` to branch by operation name and variables while keeping a single `/graphql` endpoint.

```yaml
routes:
  - path: /graphql
    method: POST
    match:
      - graphql:
          operation: "^GetUser$"
          variables:
            id: "^1$"
        response:
          data:
            user:
              id: "1"
              name: Alice
      - graphql:
          operation: "^ListUsers$"
        response:
          data:
            users:
              - id: "1"
                name: Alice
              - id: "2"
                name: Bob
    response:
      errors:
        - message: unsupported operation
```

Try it:

```sh
specter examples graphql -o config.yml -f
specter -c config.yml
curl -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"operationName":"GetUser","variables":{"id":"1"},"query":"query GetUser($id: ID!) { user(id: $id) { id name } }"}'
```

## Webhooks

Use `webhook` when the original request should return immediately but your app also expects an async callback.

```yaml
routes:
  - path: /orders
    method: POST
    status: 202
    response:
      id: ord_123
      status: accepted
    webhook:
      url: http://localhost:9999/hooks/orders
      method: POST
      delay: 250
      headers:
        X-Specter-Event: order.accepted
      body:
        order_id: ord_123
        status: accepted
```

Try it with a local callback listener in another terminal:

```sh
nc -l 9999
curl -X POST http://localhost:8080/orders -d '{}'
```

## SSE

Use `stream: true` for Server-Sent Events. Add `stream_repeat: true` when you want the sequence to loop until the client disconnects.

```yaml
routes:
  - path: /events
    method: GET
    stream: true
    stream_repeat: true
    events:
      - event: ready
        id: "1"
        data:
          status: connected
      - event: heartbeat
        id: "2"
        delay: 1000
        data:
          ok: true
```

Try it:

```sh
specter examples sse -o config.yml -f
specter -c config.yml
curl -N http://localhost:8080/events
```

## OpenAPI

The OpenAPI example generates two files: `config.yml` and `openapi.yml`. Use strict request/response validation when you want mock data to stay aligned with an API contract.

```yaml
openapi: ./openapi.yml
openapi_strict: true
openapi_strict_response: true

routes:
  - path: /pets
    method: GET
    response:
      - id: 1
        name: Fido

  - path: /pets
    method: POST
    status: 201
    response:
      id: 2
      name: "{{ .body.name }}"
```

Try it:

```sh
specter examples openapi -o config.yml -f
specter doctor -c config.yml
specter -c config.yml
curl -X POST http://localhost:8080/pets \
  -H 'Content-Type: application/json' \
  -d '{"name":"Mochi"}'
```

## Polling / Long-Running Job

Use `responses` with `mode: sequential` to model a job that moves from queued to running to complete over repeated calls.

```yaml
routes:
  - path: /jobs
    method: POST
    status: 202
    response:
      id: job_123
      status: queued

  - path: /jobs/job_123
    method: GET
    mode: sequential
    responses:
      - status: 202
        response:
          id: job_123
          status: queued
      - status: 202
        delay: 500
        response:
          id: job_123
          status: running
      - status: 200
        response:
          id: job_123
          status: complete
          result:
            download_url: https://example.com/report.csv
```

Try it:

```sh
curl -X POST http://localhost:8080/jobs
curl http://localhost:8080/jobs/job_123
curl http://localhost:8080/jobs/job_123
curl http://localhost:8080/jobs/job_123
```

## Error And Latency Scenarios

Use rate limits, injected error probability, and delays to test retry logic, loading states, and error UI.

```yaml
routes:
  - path: /unstable
    method: GET
    delay_min: 150
    delay_max: 900
    error_rate: 0.25
    error_status: 503
    response:
      ok: true

  - path: /limited
    method: GET
    rate_limit: 2
    rate_reset: 60
    response:
      ok: true

  - path: /missing
    method: GET
    status: 404
    response:
      error: not_found

  - path: /bad-request
    method: POST
    status: 400
    response:
      error: invalid_request
```

Try it:

```sh
specter examples errors -o config.yml -f
specter -c config.yml
curl http://localhost:8080/unstable
curl http://localhost:8080/limited
```

## Choosing An Example

- Start with `auth` when your app needs login gates, permissions, or session-shaped UI.
- Start with `crud` or `pagination` when you need realistic list and detail screens.
- Start with `openapi` when API contract drift is the biggest risk.
- Start with `errors` when you are hardening retries, loading states, or empty/error pages.
- Use `doctor` after generating or editing an example:

```sh
specter doctor -c config.yml
```
