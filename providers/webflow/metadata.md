# Webflow connector

Deep connector for the [Webflow Data API v2](https://developers.webflow.com/data/reference).
Authentication is OAuth 2.0 (authorization code). The catalog base URL is
`https://api.webflow.com`; object paths carry the `/v2` prefix.

## Supported objects

| Object | Resource path | Records key | Scope |
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

Object names are the last URL segment, except `custom_code_blocks` and
`google_tags`, whose last segment alone is not descriptive.

`redirects` and `activity_logs` require an Enterprise site (403 otherwise);
`orders` and `products` require Ecommerce to be enabled on the site (409 otherwise).

## Site scope

Every object except `sites` lives under a site. The site comes from the
`siteId` connection metadata input (declared in `providers/webflow.go`,
required by the connector through `common.RequireMetadata`) and is
substituted for `{site_id}` in the object path at request time. One OAuth
token can be authorised for several sites, so a connection maps to one site.

## Metadata

`ListObjectMetadata` is served from the static `metadata/schemas.json`,
generated from Webflow's official OpenAPI spec. For every object the fields
are taken from the spec's list-endpoint response schema as-is; `products` is
the one exception, where the schema is reshaped (see below) because the API
wraps each record in a `{product, skus}` envelope.

- Spec: `scripts/openapi/webflow/internal/files/v2.yml`, from
  [webflow/openapi-spec](https://github.com/webflow/openapi-spec) (MIT),
  stored with Git LFS. Run `git lfs pull` after cloning before regenerating.
- Generator: `go run ./scripts/openapi/webflow/metadata`. Objects are an
  explicit allow-list in the script.
- The spec is OpenAPI 3.1 and fully inlined. `internal/files/file.go`
  rewrites three constructs in memory so `kin-openapi` can load it: tuple
  `items` in the shared error schema become `oneOf`, item schemas with
  `properties` but no `type` get `type: object`, and map-valued `examples`
  on schemas are dropped. The file on disk stays verbatim.
- `internal/files/live-fields.json` adds fields seen in live responses but
  missing from the spec (`sites.fullSiteCompiledAt`, `pages.shouldPublish`,
  `forms.componentId`, `forms.componentElementId`).
- `products` is the only object not taken verbatim from the spec: records
  are `{product, skus}` envelopes, and the generator lifts the product's
  fields to the top level with `skus` kept as a nested array, so the object
  describes a product.

Refresh the spec with:

```sh
curl -L -o scripts/openapi/webflow/internal/files/v2.yml \
  https://raw.githubusercontent.com/webflow/openapi-spec/main/openapi/v2.yml
go run ./scripts/openapi/webflow/metadata
```

## Not supported

- **CMS collection items** (`/collections/{collection_id}/items`): fields are
  defined per collection (`fieldData`), so a static schema cannot describe
  them. They need a dynamic metadata provider exposing one object per
  collection.
- Nested lists needing a second parent id (comment replies, per-form
  submissions, SKU inventory, component and page DOM), analyze reports,
  singletons (`robots_txt`, `plan`, site custom code, `ecommerce/settings`,
  token endpoints) and workspace audit logs.

## Testing

- Unit tests: `go test ./providers/webflow/...`
- Live: `go run ./test/webflow/metadata` with a `webflow-creds.json` holding
  the OAuth token and `metadata.siteId`.
