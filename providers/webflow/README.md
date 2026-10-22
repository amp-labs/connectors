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

## Read

Implemented in `read.go` (request building and response parsing) with the
per-object table in `objects.go`, wired through `components.Reader` in
`connector.go`. All 19 objects above are readable.

### How a read works

1. **Path.** The object path comes from `metadata/schemas.json`, so Read and
   metadata can never disagree on an endpoint. `{site_id}` is replaced with
   the connection's `siteId`; `sites` has no placeholder.
2. **Pagination.** Webflow paginates with `limit`/`offset` and returns
   `pagination: {limit, offset, total}`. The first request sends
   `limit=<page size>&offset=0`; the next page is `offset + limit` while that
   is below `total`, and the token handed back is the full next-page URL, so a
   follow-up request is replayed as-is. Page size defaults to 100 and is
   capped at 100, Webflow's documented maximum, so a silently reduced page
   can never be mistaken for the last one.
   `sites`, `collections`, `webhooks`, `custom_domains`, `google_tags` and
   `redirects` ignore `limit`/`offset` and return the whole collection in one
   response without a `pagination` object (verified live for the first
   three); they are treated as single-page. `registered_scripts` is not
   documented with pagination parameters but accepts them (verified live), so
   it paginates like the rest.
3. **Records.** The records array is located by the `responseKey` stored in
   `schemas.json`. Records are returned as the API sends them; the only
   reshaping is `products`, where the `{product, skus}` envelope is flattened
   to the product's fields plus `skus`, mirroring the metadata. `Raw` keeps
   the untouched envelope.
4. **Record id.** `id` everywhere except `orders` (`orderId`), `google_tags`
   (`tagId`) and `products` (`product.id`). `custom_code_blocks` records have
   no identifier, so `Id` is left empty.
5. **Errors.** Standard status mapping. Plan-gated endpoints answer 403
   (`not_enterprise_plan_site`) and Ecommerce endpoints 409
   (`ecommerce_not_enabled`) until enabled on the site; the Webflow message
   is included in the error.

### Incremental read

No Webflow list endpoint accepts an "updated since" parameter (only CMS
collection items do, and they are out of scope). Connector-side filtering on
Since/Until is only applied when the endpoint can return records newest
first, so pagination can stop at the first record older than Since;
filtering an unsorted list would still fetch every page on every sync and
only look incremental. `comments` is the one list with a sort
(`sortBy=lastUpdated&sortOrder=desc`, verified live), so it is the only
object with incremental support: read newest-first and filtered by
`lastUpdated` with `readhelper`'s reverse-order time filter, inclusive on
both ends. Every other object ignores Since/Until and is read in full.

| Object | Incremental | Reason |
|---|---|---|
| `comments` | yes, `lastUpdated` newest first, early stop | only list with `sortBy`/`sortOrder` |
| `activity_logs`, `asset_folders`, `assets`, `custom_code_blocks`, `forms`, `pages`, `products`, `registered_scripts` | no | carry `lastUpdated` but no sort parameter (assets verified live to come back in no date order) |
| `sites`, `collections`, `webhooks`, `custom_domains`, `google_tags`, `redirects` | no | single response, no sort |
| `form_submissions` | no | `dateSubmitted` only, no sort |
| `components`, `custom_fonts` | no | no timestamp |
| `orders` | no | `acceptedOn`/`fulfilledOn`/`refundedOn`/`disputeUpdatedOn` but no last-modified field and no sort |

If Webflow adds sorting to another list, enabling incremental read for it is
the `incrementalKey` entry in `objects.go`.

### Decisions

- **One connection, one site.** The site comes from the `siteId` input rather
  than fanning out over every authorised site. It matches how Breezy
  (`company_id`) and Mailgun (domain) scope their objects, keeps record
  identity and pagination simple, and most records carry `siteId` anyway.
- **Next-page token is the full URL.** Same as BambooHR and Mailgun; the
  follow-up request needs no object-specific reconstruction.
- **Products flattened on both sides.** Metadata and Read describe the same
  shape; the raw envelope is still available on each row.
- **`Support.Read` stays off** in `providers/webflow.go` until the separate
  enable PR, following the BambooHR and Breezy sequence.

### Verified live (2026-10-06)

Against a non-Enterprise site without Ecommerce: `sites`, `pages`, `forms`
return their records with correct ids; `assets` pages through 6 records with
page sizes 2 and 4 (distinct ids, no extra empty request); empty objects
return a finished empty page; `comments` with Since set returns cleanly.
`redirects`, `activity_logs`, `orders` and `products` could not be exercised
on that site (403/409) and are covered by unit tests only.

## Write

Implemented in `write.go` with the per-object table `writeSpecs`, wired
through `components.Writer` in `connector.go`. Create is `WriteParams` without
a `RecordId`; update is with one. Objects without the matching endpoint return
`ErrOperationNotSupportedForObject`.

| Object | Create | Update |
|---|---|---|
| `collections` | `POST /v2/sites/{site_id}/collections` | `PATCH /v2/collections/{id}` |
| `assets` | `POST /v2/sites/{site_id}/assets` (registers an upload, see below) | `PATCH /v2/assets/{id}` |
| `custom_fonts` | `POST /v2/sites/{site_id}/custom_fonts` (registers an upload) | `PATCH /v2/sites/{site_id}/custom_fonts/{id}` |
| `products` | `POST /v2/sites/{site_id}/products` | `PATCH /v2/sites/{site_id}/products/{id}` |
| `redirects` | `POST /v2/sites/{site_id}/redirects` | `PATCH /v2/sites/{site_id}/redirects/{id}` |
| `asset_folders` | `POST /v2/sites/{site_id}/asset_folders` | none in the API |
| `webhooks` | `POST /v2/sites/{site_id}/webhooks` | none in the API |
| `registered_scripts` | `POST …/registered_scripts/inline` or `…/hosted` | none in the API |
| `sites` | none (needs a workspace id) | `PATCH /v2/sites/{id}` |
| `pages` | none | `PUT /v2/pages/{id}` |
| `form_submissions` | none | `PATCH /v2/sites/{site_id}/form_submissions/{id}` |
| `comments` | none | `PATCH /v2/sites/{site_id}/comments/{id}` (body `{resolved}`) |
| `orders` | none | `PATCH /v2/sites/{site_id}/orders/{id}` |

Not writable: `activity_logs`, `components`, `custom_code_blocks`,
`custom_domains`, `forms`, `google_tags` (no per-record write endpoint;
Google tags are replaced as a whole list).

### How a write works

- **Payload.** `RecordData` is sent as the JSON body unchanged. Webflow
  ignores unknown and read-only fields on PATCH (verified live on assets with
  `id`, `createdOn`, `siteId` and a bogus field, all 200) but validates the
  types of known fields strictly, so a page record copied from Read with
  `null` values is rejected with `validation_error`; send only the fields to
  change.
- **Paths** are explicit per object because several update endpoints are not
  under the site (`collections`, `pages`, `assets`, `sites`). Updates use
  PATCH except `pages` (PUT).
- **Record id** in the result is `id` from the response, `orderId` for
  orders, `product.id` for a product create (the response is `{product,
  skus}`) and `customFont.id` for custom fonts. When the response has no id
  the request's `RecordId` is echoed back.
- **Registered scripts** have two create endpoints. A payload with
  `sourceCode` is registered as an inline script, anything else as a hosted
  one (`hostedLocation` + `integrityHash`).
- **Uploads.** Creating an asset or custom font registers the upload:
  Webflow returns the new id plus `uploadUrl`/`uploadDetails` (assets) or
  `upload` (fonts), and the caller must then PUT the file to that URL. The
  connector does not transfer file bytes.
- **Products** accept the flattened shape Read returns: product fields at
  the top level, SKUs under `skus`, optional `publishStatus`. The connector
  builds the API's `{product, sku, publishStatus}` body from it, using `sku`
  if given or else the first element of `skus` (the API writes one SKU per
  call; create requires one, update accepts none). Any further entries in
  `skus` are dropped: additional SKUs are created through
  `POST /v2/sites/{site_id}/products/{product_id}/skus`, which is not an
  object in this connector. A payload that already has a `product` key is
  taken as the API shape and sent unchanged. Not verified live (the test site
  has no Ecommerce); covered by unit tests.

### Verified live (2026-10-06)

Webhook create, collection create and update, and page title update (then
reverted) succeeded against the test site. `redirects` and site update are
Enterprise-gated (403) and `orders`/`products` need Ecommerce (409), so those
paths are covered by unit tests only.

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
- Live: `go run ./test/webflow/metadata`, `go run ./test/webflow/read` and
  `go run ./test/webflow/write` with a `webflow-creds.json` holding the OAuth
  token and `metadata.siteId`. The write script leaves a webhook and a
  collection behind until delete support lands.
