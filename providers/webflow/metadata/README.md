# Webflow Metadata

`schemas.json` is generated from the official Webflow Data API v2 OpenAPI spec
by `scripts/openapi/webflow/metadata`. Do not edit it by hand.

The spec lives at `scripts/openapi/webflow/internal/files/v2.yml` and is
stored with **Git LFS** (registered in `.gitattributes`). After a fresh clone
run `git lfs pull` before regenerating.

## Refreshing the spec

Webflow publishes the spec on GitHub (MIT licensed):

```sh
curl -L -o scripts/openapi/webflow/internal/files/v2.yml \
  https://raw.githubusercontent.com/webflow/openapi-spec/main/openapi/v2.yml
go run ./scripts/openapi/webflow/metadata
```

The spec is OpenAPI 3.1 and fully inlined (no `components/schemas`). Three 3.1
constructs are rewritten in memory by `scripts/openapi/webflow/internal/files/file.go`
before `kin-openapi` loads it; the file on disk stays verbatim:

- tuple `items: [{type: string}, {type: object}]` in the shared error schema
  becomes `items: {oneOf: [...]}`;
- item schemas that declare `properties` without `type: object` get the type;
- map-valued `examples` on schema nodes are dropped.

## Paths and the site scope

The catalog base URL is `https://api.webflow.com` (unversioned, because the
proxy must also serve v1), so every object path carries the `/v2` prefix.
`FlushSchemas` is used instead of `SaveSchemas` so the module path stays empty
and object paths are complete.

Every object except `sites` is site-scoped: its path contains `{site_id}`,
kept verbatim (the same approach as Mailgun's `{domain_name}` and Breezy's
`{company_id}`). The site comes from the `siteId` connection metadata input
declared in `providers/webflow.go` and required by the connector
(`common.RequireMetadata`); the read connector substitutes it into the path.
One OAuth token can be authorised for several sites (`GET /v2/token/introspect`
lists them under `authorization.authorizedTo.siteIds`), so one connection maps
to one site. Auto-resolving the site after OAuth (a `PostAuthentication`
metadata item, like Atlassian's `cloudId`) is possible later when a token is
authorised for exactly one site.

## Objects

19 objects. Names are the last URL segment (what Webflow's docs call the
resource), except `custom_code_blocks` (`/custom_code/blocks`, "blocks" alone is
meaningless) and `google_tags` (`/integrations/google_tags`).

| Object | Path | Records key | Scope |
|---|---|---|---|
| `sites` | `/v2/sites` | `sites` | `sites:read` |
| `activity_logs` | `/v2/sites/{site_id}/activity_logs` | `items` | `site_activity:read` |
| `asset_folders` | `/v2/sites/{site_id}/asset_folders` | `assetFolders` | `assets:read` |
| `assets` | `/v2/sites/{site_id}/assets` | `assets` | `assets:read` |
| `collections` | `/v2/sites/{site_id}/collections` | `collections` | `cms:read` |
| `comments` | `/v2/sites/{site_id}/comments` | `comments` | `comments:read` |
| `components` | `/v2/sites/{site_id}/components` | `components` | `components:read` |
| `custom_code_blocks` | `/v2/sites/{site_id}/custom_code/blocks` | `blocks` | `custom_code:read` |
| `custom_domains` | `/v2/sites/{site_id}/custom_domains` | `customDomains` | `sites:read` |
| `custom_fonts` | `/v2/sites/{site_id}/custom_fonts` | `customFonts` | `sites:read` |
| `form_submissions` | `/v2/sites/{site_id}/form_submissions` | `formSubmissions` | `forms:read` |
| `forms` | `/v2/sites/{site_id}/forms` | `forms` | `forms:read` |
| `google_tags` | `/v2/sites/{site_id}/integrations/google_tags` | `googleTagIds` | `sites:read` |
| `orders` | `/v2/sites/{site_id}/orders` | `orders` | `ecommerce:read` |
| `pages` | `/v2/sites/{site_id}/pages` | `pages` | `pages:read` |
| `products` | `/v2/sites/{site_id}/products` | `items` | `ecommerce:read` |
| `redirects` | `/v2/sites/{site_id}/redirects` | `redirects` | `sites:read` |
| `registered_scripts` | `/v2/sites/{site_id}/registered_scripts` | `registeredScripts` | `custom_code:read` |
| `webhooks` | `/v2/sites/{site_id}/webhooks` | `webhooks` | `sites:read` |

`products` records are `{product, skus}` pairs in the API; the generator
flattens `product` so the object describes a product (a CMS item: `id`,
`fieldData`, `isDraft`, …) with `skus` as a nested array field.

Scope names come from the spec's per-operation `security` entries. The docs
spell the page scope `pages:read` while the spec says `page:read`; verify
against a live token before relying on either.

## Live verification (2026-10-06)

Checked against a workspace-level OAuth token and a freshly created site:

- `ListObjectMetadata` returns every object with no errors.
- All 18 site-scoped endpoints plus `/v2/sites` were called; every response
  key in the table above matches the live payload, and `pagination` is present
  exactly where noted below.
- Four fields appear live but not in the spec; they are merged by the
  generator from `scripts/openapi/webflow/internal/files/live-fields.json`:
  `sites.fullSiteCompiledAt`, `pages.shouldPublish`, `forms.componentId`,
  `forms.componentElementId`.
- Plan-gated endpoints: `redirects` and `activity_logs` answer
  `403 not_enterprise_plan_site` on a non-Enterprise site; `orders` and
  `products` answer `409 ecommerce_not_enabled` until Ecommerce is turned on.
  Metadata is still served for them; the read connector should surface these
  as clear errors rather than retrying.

## Excluded endpoints and why

- **CMS collection items** (`/collections/{collection_id}/items`, `/items/live`).
  The most valuable Webflow data, but the record fields live in `fieldData` and
  are defined per collection (`GET /collections/{id}` → `fields[]` with types
  such as PlainText, RichText, Image, Reference, Option, DateTime, Switch). A
  static schema cannot describe them; they need a dynamic metadata provider
  that lists collections and exposes one object per collection. Planned as a
  follow-up, not part of this change.
- **Two-level nesting** needing a parent id the connector API does not carry:
  comment replies (`/comments/{thread}/replies`), per-form submissions
  (`/forms/{form_id}/submissions`, redundant with the site-level list), SKU
  inventory, component DOM/properties, page DOM.
- **Analyze reports** (`/analyze/reports/*`): computed aggregates with required
  `startTime`/`endTime`, not records.
- **Singletons / envelopes**: site custom code, `robots_txt`, `plan`,
  `ecommerce/settings`, `token/introspect`, `token/authorized_by`.
- **Workspace audit logs** (`/workspaces/{id}/audit_logs`): needs a workspace
  id (also in the introspect response); can be added once scope resolution is
  settled.

## Read-phase notes (not implemented here)

- Pagination is `offset`/`limit` with `pagination: {limit, offset, total}`;
  `limit` max is 100. `sites`, `collections`, `custom_domains`, `google_tags`,
  `webhooks` and `redirects` return the full list without a pagination object
  (verified live for all but `redirects`).
- Most records carry `lastUpdated` (`createdOn` where immutable); form
  submissions use `dateSubmitted`, orders `acceptedOn`. The list endpoints
  accept no `since` filter except collection items, so incremental sync will
  be client-side filtering.
- Rate limits: 60 or 120 requests/minute per token depending on plan; 429 with
  `Retry-After` and `X-RateLimit-Remaining` headers.
