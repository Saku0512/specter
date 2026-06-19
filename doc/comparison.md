# Comparison: Specter, json-server, Prism, and WireMock

This page helps you choose between Specter and other common mock-server tools. It is not a scoreboard: each tool has a different center of gravity.

- Choose **Specter** when you want one local YAML config that combines route mocks, request matching, state, stores, scenarios, timelines, OpenAPI validation, request assertions, and a small Web UI.
- Choose **json-server** when you mainly need an instant REST API over a JSON database.
- Choose **Prism** when your OpenAPI or Postman contract is the primary source of truth and you want spec-derived mocks or validation proxy behavior.
- Choose **WireMock** when you need very rich request matching, verification, JVM/test-framework integration, record/playback, or a mature service virtualization ecosystem.

## Quick Matrix

| Topic | Specter | json-server | Prism | WireMock |
|---|---|---|---|---|
| Config style | YAML routes and scenarios | JSON/JSON5 database | OpenAPI/Postman document first | JSON mappings or Java/JVM API |
| Learning curve | Small for YAML-first mocks; grows with state/scenarios | Very small for CRUD data | Small if you already have OpenAPI | Moderate; powerful but more concepts |
| OpenAPI support | Optional request/response validation and config generation | Not the core model | Core model for mocking and validation proxy | Not the core model; usually handled through adjacent workflows or extensions |
| Dynamic state/stores | Built-in state, vars, named stores, and CRUD route wiring | Database is stateful as data changes | Primarily contract-derived; persistence/sandbox behavior is not the center | Stateful behavior via scenarios and rich stubs |
| Scenarios/timelines | Named scenarios plus route timelines and reset endpoints | Not a main concept | Not a main concept | Stateful scenarios are a core capability |
| Request inspection/assertions | Built-in request history and `/__specter/requests/assert` | Not a primary feature | Useful for contract validation, not request assertion history | Strong verification APIs |
| Web UI | Built-in local dashboard for requests, routes, state, stores, timelines, and config validation | No comparable built-in dashboard | Stoplight hosted/design tools exist; CLI Prism itself is server-focused | WireMock Cloud and ecosystem tools; OSS standalone is API/admin focused |
| Local dev / E2E workflow | Designed for frontend dev and E2E setup/teardown with reset/scenario/assert APIs | Great for quick CRUD prototypes | Great for contract-first mocks in CI and local dev | Strong for integration testing and service virtualization |

## Where Specter Fits Best

Specter is most useful when the mock needs to behave like a small local system, not only a static endpoint collection.

Good fits:

- Frontend development where each screen needs different backend states.
- E2E tests that need setup/teardown through HTTP APIs.
- Demo environments where non-developers need to switch scenarios quickly.
- API contract checks where OpenAPI validation is useful but not the only source of behavior.
- Mock APIs that need request history, assertions, dynamic routes, webhooks, SSE, GraphQL matching, latency, flaky responses, and CRUD stores in one config.

Tradeoffs:

- Specter is newer and smaller than WireMock's ecosystem.
- It is YAML-first, so teams already standardized on Java/JVM test fixtures may prefer WireMock.
- It can use OpenAPI, but Prism is more directly spec-first.
- It can create CRUD-like APIs, but json-server is still the fastest option when a JSON database is the entire mock.

## Specter vs. json-server

json-server is excellent when your mock can be represented as data collections. A `db.json` quickly becomes REST endpoints for arrays and objects, with built-in filtering, sorting, pagination, and relationship helpers.

Specter is a better fit when:

- You need multiple responses for the same method/path based on headers, query, body, cookies, form data, or GraphQL variables.
- You need login state, scenario presets, timelines, rate limits, webhooks, SSE, or flaky responses.
- You want request history and assertion endpoints for E2E tests.
- You want a Web UI to inspect requests, state, stores, and route progress while the app runs.

json-server is a better fit when:

- Your mock is mostly a REST wrapper over a JSON file.
- You want the smallest possible setup for CRUD list/detail/create/update/delete screens.
- You do not need scenario resets, timeline progress, or request assertions.

Migration shape:

```yaml
stores:
  posts:
    seed:
      - id: "1"
        title: Hello

routes:
  - path: /posts
    method: GET
    store_list: posts

  - path: /posts
    method: POST
    store_push: posts
```

## Specter vs. Prism

Prism is built around API contracts. It turns OpenAPI v2/v3 and Postman Collection files into mock servers and can also proxy traffic to validate API behavior against a spec.

Specter is a better fit when:

- You want to hand-author behavior that goes beyond examples in an OpenAPI document.
- You need stateful UI flows such as login, checkout, polling jobs, or scenario presets.
- You want in-memory stores, request assertions, and dynamic test setup APIs.
- You want a local Web UI alongside OpenAPI request/response validation.

Prism is a better fit when:

- Your API spec is the source of truth and examples/schemas should drive mock responses.
- You need a validation proxy between a client and a real backend.
- Your team already works in a contract-first OpenAPI workflow.

Migration shape:

```yaml
openapi: ./openapi.yml
openapi_strict: true
openapi_strict_response: true

routes:
  - path: /pets
    method: POST
    status: 201
    response:
      id: 2
      name: "{{ .body.name }}"
```

## Specter vs. WireMock

WireMock is a mature service virtualization tool with rich matching, response templating, stateful behavior, verification, record/playback, Java/JVM integrations, admin APIs, and commercial/cloud options.

Specter is a better fit when:

- You want a small single-binary local tool with a YAML config and no JVM dependency.
- Frontend developers need to read and edit mock behavior directly.
- You want built-in scenario switching, stores, request assertions, dynamic routes, and a Web UI without assembling a larger test harness.
- Your team wants a lightweight tool for local dev and E2E rather than a broad service virtualization platform.

WireMock is a better fit when:

- You need advanced matching and verification across many protocols and extensions.
- Your tests already run in Java/JVM or use Testcontainers/Spring Boot integrations.
- You need record/playback, custom matchers, extension points, or WireMock Cloud.
- You are virtualizing many services across a larger organization.

Migration shape:

```yaml
routes:
  - path: /checkout
    method: POST
    match:
      - headers:
          X-Test-Case: "^approved$"
        response:
          status: approved
      - body_path:
          payment.amount: "^0$"
        status: 400
        response:
          error: invalid_amount
    status: 402
    response:
      error: payment_required
```

## Topic Notes

### Config Style And Learning Curve

Specter uses one YAML file with top-level settings, routes, stores, and scenarios. That keeps local mocks readable for frontend and QA workflows. json-server uses a database-shaped file, which is even simpler when the API is mostly CRUD. Prism asks you to think in OpenAPI/Postman first. WireMock is extremely expressive, but teams usually need to learn its mapping model, admin APIs, and verification concepts.

### OpenAPI Support

Specter can generate a starter config from OpenAPI and validate requests/responses at runtime. Prism is more OpenAPI-native: the spec directly drives the mock server and validation proxy. WireMock and json-server can still participate in contract workflows, but OpenAPI is not their simplest core model.

### Dynamic State, Stores, Scenarios, And Timelines

Specter has first-class state, vars, named stores, scenario presets, and timeline progress reset endpoints. This makes it convenient for E2E tests that need known states before each test. WireMock also has stateful behavior through scenarios. json-server's state is mostly the mutable database. Prism is strongest when behavior is derived from the contract rather than custom state machines.

### Request Inspection And Assertions

Specter records incoming requests and exposes assertion endpoints such as `POST /__specter/requests/assert`. This keeps tests language-agnostic: Playwright, Cypress, shell scripts, or any HTTP client can verify that an expected request happened. WireMock has strong verification APIs. Prism focuses more on contract validation. json-server is not primarily a request assertion tool.

### Web UI Capabilities

Specter's local dashboard shows recorded requests, routes, state, vars, stores, timelines, and config validation. This is aimed at day-to-day local debugging. WireMock has admin APIs and a broader ecosystem, including WireMock Cloud. Prism has hosted/design workflows through Stoplight, while the open-source CLI server is primarily command-line/server focused. json-server keeps the surface intentionally small.

## Decision Guide

Use Specter if you say:

- "I need mocks that change state during an E2E test."
- "I want a config my frontend team can read without writing JavaScript or Java."
- "I need request assertions but my tests should stay language-agnostic."
- "I want OpenAPI validation plus custom mock behavior."
- "I need a local dashboard to see what the app just called."

Use another tool if you say:

- "I only need CRUD over JSON data." Use json-server.
- "Everything should be generated from OpenAPI." Use Prism.
- "I need enterprise-grade service virtualization or deep JVM integration." Use WireMock.

## References

- [json-server](https://github.com/typicode/json-server)
- [Prism](https://github.com/stoplightio/prism)
- [WireMock request matching](https://wiremock.org/docs/request-matching/)
- [WireMock stateful behaviour](https://wiremock.org/docs/stateful-behaviour/)
- [WireMock verifying](https://wiremock.org/docs/verifying/)
