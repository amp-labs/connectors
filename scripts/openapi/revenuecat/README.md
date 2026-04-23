# RevenueCat OpenAPI (`swagger.yaml`)

This folder contains the RevenueCat API v2 OpenAPI specification used by our tooling to generate/refresh metadata and schemas for the RevenueCat connector.

## Source

`swagger.yaml` was downloaded from RevenueCat’s public docs:

- `https://www.revenuecat.com/docs/api-v2`

## Why we “clean” the file

The upstream spec includes large `examples:` blocks. Some of those examples contain secret-looking strings (for example `shared_secret` values and private key PEM blocks). Our CI runs Semgrep secret-detection rules, and those example literals can trigger blocking findings even though they are documentation samples.

We don’t rely on OpenAPI `examples:` for schema extraction, so we remove them to:

- Prevent secret-scanner false positives in CI
- Reduce file size/noise in diffs

## Cleaning steps applied

When updating `swagger.yaml` from the docs, apply these steps before committing:

- Remove **all** OpenAPI `examples:` blocks from the YAML.
- Ensure no PEM key blocks are present (e.g. `-----BEGIN … PRIVATE KEY-----`).
- Keep the actual schema definitions intact (we still rely on them for metadata extraction).

## Regenerating connector metadata

This directory includes a generator (`main.go`) which embeds `swagger.yaml` and extracts top-level GET endpoints into the RevenueCat metadata package.

From the repo root:

```bash
go run ./scripts/openapi/revenuecat
```

