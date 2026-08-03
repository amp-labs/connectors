# Mailgun Metadata

`schemas.json` is auto-generated from the Mailgun OpenAPI specification by
`scripts/openapi/mailgun/metadata`. Do not edit it by hand.

The OpenAPI spec lives at `openapi/mailgun.yaml` in this directory. It is
OpenAPI 3.1.0 and declares both regional servers (US `https://api.mailgun.net`
and EU `https://api.eu.mailgun.net`).

## Refreshing the spec

Mailgun does not link a downloadable spec from its developer site, but the
Redocly docs portal serves it from a `_spec` path (the same URL Mailgun's own
SDK build uses). To refresh:

1. Download the latest spec:

   ```sh
   curl -o providers/mailgun/metadata/openapi/mailgun.yaml \
     "https://documentation.mailgun.com/_spec/docs/mailgun/api-reference/send/mailgun.yaml?download"
   ```

2. From the repo root, run:

   ```sh
   go run ./scripts/openapi/mailgun/metadata
   ```

The generator projects the spec down to the connector's supported object
inventory (defined in `getEndpoints` / `postEndpoints` in the script). To add
or remove an object, edit those maps and re-run the generator.

## Why some endpoints aren't here

The spec is a merged, multi-version (v1..v5) bundle of several internal Mailgun
services with ~169 paths; we expose 27 as connector objects. Object names are
curated (rather than auto-derived from the last path segment) because
auto-naming collides across versions and leaks path parameters.

Excluded categories:

- **Get-by-id singletons** (`/v3/routes/{id}`, `/v4/domains/{name}`,
  `/v5/users/{user_id}`, `.../bounces/{address}`, …) — no list semantics; the
  read layer fetches these by id.
- **Aggregated analytics** — both the legacy GET `/v3/stats/*`,
  `/v3/{domain}/aggregates/*`, `/v3/{domain}/tag/stats`, `/v1/thresholds/hits`,
  `/v1/bounce-classification/stats` and the modern POST `/v1/analytics/metrics`.
  These return computed totals, not records ("expose data, not endpoints").
- **Action endpoints** (`/v3/routes/match`, `/v3/ips/request/new`,
  `/v3/ips/account/settings`, `/v5/accounts/http_signing_key`, …).
- **Reference lookups** (`/v3/domains/{domain}/tag/{countries,devices,providers}`).
- **Pagination-variant duplicates** (`/v3/lists/pages`,
  `/v3/lists/{list_address}/members/pages`).
- **Sandbox-only** (`/v5/sandbox/auth_recipients`).
- **Deprecated** GET `/v3/{domain_name}/events` — superseded by the modern
  Logs API (POST `/v1/analytics/logs`), which is exposed as the `logs` object.

## Domain scope

Most objects are domain-scoped: their path contains a domain placeholder that
appears under several names in the spec (`{domain_name}`, `{domain}`, `{name}`,
`{authority_name}`). The placeholder is kept verbatim in `schemas.json`; the
read connector substitutes it from connector metadata (the Mailgun domain).
