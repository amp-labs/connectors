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
services with 169 paths (265 operations); we expose 30 as connector objects.
Object names are the URL path with the version and any domain placeholder
removed, prefixes preserved. `account/templates` is the one curated name:
`/v4/templates` would collide with the domain-scoped `/v3/{domain_name}/templates`
(both truncate to `templates`), and the prefix mirrors Mailgun's own "Account
Templates" / "Domain Templates" docs sections.

`alerts/settings` is read from the `events` array of the `/v1/alerts/settings`
envelope and created, updated and deleted at `/v1/alerts/settings/events`; one
name covers both, per the "same object across operations" rule.

Excluded categories:

- **Get-by-id singletons** (`/v3/routes/{id}`, `/v4/domains/{name}`,
  `/v5/users/{user_id}`, `.../bounces/{address}`, …) — no list semantics.
- **Aggregated analytics** — both the legacy GET `/v3/stats/*`,
  `/v3/{domain}/aggregates/*`, `/v3/{domain}/tag/stats`,
  `/v1/bounce-classification/*` and the modern POST `/v1/analytics/metrics`,
  `/v1/analytics/usage/metrics`, `/v2/bounce-classification/metrics`.
  These return computed totals, not records ("expose data, not endpoints").
- **Action endpoints** (`/v3/routes/match`, `/v3/ips/request/new`,
  `/v3/ips/account/settings`, `/v5/accounts/http_signing_key`, …).
- **Reference lookups** (`/v3/domains/{domain}/tag/{countries,devices,providers}`,
  `/v1/alerts/events`, `/v1/alerts/slack/channels`).
- **Pagination-variant duplicates** (`/v3/lists/pages`,
  `/v3/lists/{list_address}/members/pages`).
- **Representation variants** — `/v3/ips/details/all` lists the same IPs as
  `/v3/ips` with account bookkeeping columns.
- **Two-level nesting** — `/v3/{domain_name}/templates/{name}/versions` and
  `/v4/templates/{name}/versions` need a parent id the connector API does not
  carry (only `lists/members` is fanned out, over the account's lists).
- **Legacy domain webhooks** — `/v3/domains/{domain}/webhooks` returns a map
  keyed by event name, not a list; account webhooks (`/v1/webhooks`) cover it.
- **Spec gap** — `/v3/ip_warmups` is documented as returning `{message}` but
  the live API returns `items` + `paging`; with no schema to derive fields
  from, it is left out.
- **Sandbox-only** (`/v5/sandbox/auth_recipients`).
- **Deprecated** GET `/v3/{domain_name}/events` — superseded by the modern
  Logs API (POST `/v1/analytics/logs`), exposed as the `analytics/logs` object.
- **Deprecated** GET `/v3/{domain}/tags` — superseded by the modern,
  account-scoped Tags API (POST `/v1/analytics/tags`), exposed as the
  `analytics/tags` object.

## Known spec discrepancies (types)

`schemas.json` reflects the spec verbatim, so two spec issues surface in the
field types. Neither affects the values Read returns; only the advertised type.

- The suppression timestamps — `created_at` on `bounces`, `complaints` and
  `unsubscribes`, `createdAt` on `whitelists` — are declared `type: object`
  with an empty `allOf`. The live API returns RFC 1123 strings (e.g.
  `Thu, 11 Dec 2025 01:49:40 UTC`), which the connector's `Since`/`Until`
  filtering relies on. They therefore appear as `other` instead of `string`.
- Fields declared as OpenAPI 3.1 unions (`oneOf: [string, null]`, `anyOf`
  objects), e.g. `keys.domain_name`, `users.github_user_id`,
  `domains/credentials.size_bytes`, carry an empty provider type: the shared
  `api3` converter maps only `integer`, `boolean` and `string`.

## Domain scope

Seven objects are domain-scoped (`bounces`, `complaints`, `unsubscribes`,
`whitelists`, `templates`, `domains/credentials`, `domains/keys`): their path
contains a domain placeholder, spelled `{domain_name}` in the spec (or
`{authority_name}` for `domains/keys`). The placeholder is kept verbatim in
`schemas.json`; the connector substitutes the workspace (the Mailgun sending
domain) at request time for read, write and delete.
