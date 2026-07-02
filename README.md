# specter

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Saku0512/specter)
[![CI](https://github.com/Saku0512/specter/actions/workflows/test.yml/badge.svg)](https://github.com/Saku0512/specter/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Saku0512/specter)](https://goreportcard.com/report/github.com/Saku0512/specter)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/Saku0512/specter/badge)](https://scorecard.dev/viewer/?uri=github.com/Saku0512/specter)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Saku0512/specter)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/Saku0512/specter)](https://github.com/Saku0512/specter/releases/latest)
[![VS Code Marketplace](https://img.shields.io/visual-studio-marketplace/v/Saku0512-sec.specter-mock-server?label=VS%20Code%20Extension)](https://marketplace.visualstudio.com/items?itemName=Saku0512-sec.specter-mock-server)
[![VS Code Installs](https://img.shields.io/visual-studio-marketplace/i/Saku0512-sec.specter-mock-server)](https://marketplace.visualstudio.com/items?itemName=Saku0512-sec.specter-mock-server)
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

## リリース検証

リリースには `SHA256SUMS.txt` と SPDX JSON SBOM が添付されます。`install.sh` と `install.ps1` は、バイナリを配置する前に SHA256 を検証します。

手動で確認する場合:

```sh
VERSION=v1.0.1
ASSET=specter_linux_amd64
curl -LO "https://github.com/Saku0512/specter/releases/download/${VERSION}/${ASSET}"
curl -LO "https://github.com/Saku0512/specter/releases/download/${VERSION}/SHA256SUMS.txt"
grep "  ${ASSET}$" SHA256SUMS.txt | sha256sum --check -
gh attestation verify "${ASSET}" --repo Saku0512/specter
```

## Quick start

```sh
specter init          # config.yml を生成
specter init --template crud  # CRUD スターターを生成
specter examples      # サンプル設定を一覧表示
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
- [Examples gallery](doc/examples.md) — auth、CRUD、pagination、GraphQL、webhook、SSE、OpenAPI など
- [Comparison guide](doc/comparison.md) — json-server、Prism、WireMock との違い
- [JSON Schema](doc/schema.md) — エディタ補完・validation 用 schema
- [VS Code Extension](https://marketplace.visualstudio.com/items?itemName=Saku0512-sec.specter-mock-server) — Specter config の補完と inline validation
- [CLI reference](doc/cli.md) — flags、env vars、`init` / `examples` / `gen` / `validate` / `doctor` / `record`
- [Introspection API](doc/introspection.md) — `/__specter/requests`、`/__specter/state`

動作する設定例は [config.example.yml](config.example.yml) を見てください。

## 開発

通常のテストは `make test` で実行できます。Go fuzzing をローカルで回す場合は次を使います。

```sh
make fuzz
FUZZTIME=30s make fuzz
```

fuzz target は config YAML parser と request matcher をインメモリで検証するため、ネットワークアクセスは不要です。

## ライセンス

MIT
