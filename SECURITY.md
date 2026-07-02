# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | ✅ |

## Reporting a Vulnerability

Please **do not** report security vulnerabilities through public GitHub Issues.

Instead, report them via GitHub's private vulnerability reporting:
**[Report a vulnerability](https://github.com/Saku0512/specter/security/advisories/new)**

Include the following in your report:

- Description of the vulnerability
- Steps to reproduce
- Potential impact

You will receive a response within **7 days**. If the issue is confirmed, a patch will be released as soon as possible.

## Scope

This project is a **local development tool** intended to run on trusted networks. It is not designed for production use or exposure to the public internet.

- Do not expose specter to the public internet without additional security measures (firewall, reverse proxy, auth, etc.)
- Configuration files may contain sensitive response data — handle them accordingly

## Supply Chain Controls

The project uses several automated checks to reduce supply chain risk:

- Dependabot monitors GitHub Actions, Go modules, npm workspaces, and the Dockerfile.
- Pull requests run dependency review and fail on newly introduced high-severity runtime vulnerabilities.
- Go modules are verified with `go mod verify` and scanned with `govulncheck`.
- npm workspaces run `npm audit --audit-level=high`.
- OpenSSF Scorecard runs on the default branch and uploads SARIF results to code scanning.
- Release binaries include SHA256 checksums, SBOMs, and GitHub artifact attestations.
- Container images are built with BuildKit SBOM and provenance attestations.
- The Docker builder image is pinned by digest and monitored by Dependabot's Docker ecosystem updates.

## Dependency Advisory Triage

Dependency advisories are triaged with multiple scanners before accepting risk:

- `govulncheck ./...` is used for Go reachability analysis.
- `npm audit --audit-level=high` is run for both the `site` and `vscode-extension` workspaces in CI.
- Dependabot alerts are reviewed against local scan results and patched when a practical update or override is available.

For the July 2026 Scorecard `Vulnerabilities` alert, the reachable Go vulnerability count was zero. The vulnerable `golang.org/x/*` modules were still updated to patched versions, and low-severity npm advisories in transitive development dependencies were resolved with package overrides. No dependency advisories are intentionally accepted after this triage; future remaining alerts should be documented here with the advisory ID, affected manifest, reachability, and rationale.

## Docker Image Digest Updates

The Dockerfile pins the Go builder image as `golang:1.26.4-alpine@sha256:<digest>` so builds keep the intended tag while also locking the immutable manifest. Dependabot is configured for the Dockerfile and should open updates when the pinned digest changes.

Before merging a digest update, verify the tag-to-digest mapping and build locally:

```sh
docker buildx imagetools inspect golang:1.26.4-alpine
docker build -t specter:local .
```

The inspected index digest should match the Dockerfile digest, and the manifest annotations should still identify the intended Go/Alpine image.

## Release Workflow Token Permissions

The release workflow defaults to `contents: read`. Write-scoped tokens are limited to the jobs that publish release outputs:

- `release` uses `contents: write` to create the GitHub Release and upload release assets. It also uses `id-token: write`, `attestations: write`, and `artifact-metadata: write` to generate GitHub artifact attestations.
- `prepare-formula-update` uses only `contents: read` while downloading release assets, calculating SHA256 checksums, and preparing the Homebrew formula diff.
- `open-formula-pr` uses `contents: write` only to push the generated `chore/update-formula-*` branch, and `pull-requests: write` to open the formula update PR.

The remaining write permissions are intentionally accepted because release publication and automated formula PR creation require repository writes. Read-only preparation work stays in a separate job so those write scopes are not available while checksums are computed.

For stronger CI isolation, consider enabling Harden Runner in audit mode first, then moving release and Docker workflows to an egress allowlist once expected network destinations are known.
