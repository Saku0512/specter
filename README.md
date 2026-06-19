# specter

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Saku0512/specter)
[![CI](https://github.com/Saku0512/specter/actions/workflows/test.yml/badge.svg)](https://github.com/Saku0512/specter/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Saku0512/specter)](https://goreportcard.com/report/github.com/Saku0512/specter)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Saku0512/specter)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/Saku0512/specter)](https://github.com/Saku0512/specter/releases/latest)
![Downloads](https://img.shields.io/github/downloads/Saku0512/specter/total)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[English](README-en.md)

specter は軽量なモック API サーバーです。エンドポイントを YAML で定義して、すぐに起動できます。

- Hot reload: `config.yml` を編集すると変更がすぐ反映されます
- Response templates、faker、stateful mocking、rate limiting に対応
- 単一バイナリで動作し、追加のランタイム依存はありません

## インストール

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

## Quick start

```sh
specter init          # config.yml を生成
specter init --template crud  # CRUD スターターを生成
specter doctor -c config.yml  # 設定・参照ファイル・ポートを診断
specter -c config.yml # サーバーを起動
```

```yaml
routes:
  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice
```

この設定で `GET /users` にアクセスすると、YAML に書いたレスポンスが返ります。

## ドキュメント

- [Config reference](doc/config.md) — routes、matching、templates、faker、state、rate limiting など
- [JSON Schema](doc/schema.md) — エディタ補完・validation 用 schema
- [CLI reference](doc/cli.md) — flags、env vars、`init` / `gen` / `validate` / `doctor` / `record`
- [Introspection API](doc/introspection.md) — `/__specter/requests`、`/__specter/state`

動作する設定例は [config.example.yml](config.example.yml) を見てください。

## ライセンス

MIT
