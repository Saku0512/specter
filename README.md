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
| `-v`, `--version` | — | Show version |

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

Both `.yaml` and `.yml` extensions are supported.

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

### Hot Reload

specter watches the config file and reloads automatically on save. No restart required.

```
[GIN] ...  👻 config reloaded
```

## License

MIT
