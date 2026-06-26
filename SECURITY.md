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

For stronger CI isolation, consider enabling Harden Runner in audit mode first, then moving release and Docker workflows to an egress allowlist once expected network destinations are known.
