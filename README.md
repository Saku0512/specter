# specter

Lightweight mock API server. Define endpoints in YAML, run instantly.

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

## License

MIT
