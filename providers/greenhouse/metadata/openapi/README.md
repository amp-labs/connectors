# Greenhouse Harvest v3 OpenAPI spec

`harvest-api.json` is the OpenAPI 3.1 description of the Greenhouse Harvest **v3** API. It is the
input consumed by `scripts/openapi/greenhouse/metadata` to produce
`providers/greenhouse/metadata/schemas.json`.

## Which API this describes

Greenhouse runs two Harvest APIs:

| Version | Docs | Pagination | Incremental filter |
| ------- | ---- | ---------- | ------------------ |
| v1/v2 (deprecated, removed 2026-08-31) | https://developers.greenhouse.io/harvest.html | `page` / `per_page` | `updated_after` |
| **v3 (this connector)** | https://harvestdocs.greenhouse.io | `cursor` + `Link` header | `updated_at[gte]` |

The two are easy to confuse because they share the `https://harvest.greenhouse.io` host and much of
their object vocabulary. When checking behaviour, confirm the page is under `harvestdocs.greenhouse.io`.

## Where the spec comes from

**Greenhouse does not publish a downloadable OpenAPI file.** There is no spec in their public docs
repository (`grnhse/greenhouse-api-docs`, which covers v1), `harvestdocs.greenhouse.io/llms.txt`
links to none, and the usual hosted-docs download paths all return 404.

The documentation site is ReadMe-hosted, and every page under `/reference/` embeds the complete
spec inline in the served HTML. It is present in the raw document, so no JavaScript execution is
needed to retrieve it.

To refresh:

1. Fetch any reference page, e.g. `https://harvestdocs.greenhouse.io/reference/get_v3-candidates`.
2. Locate the first occurrence of `{"openapi":"3.1.0"` and brace-match to the closing `}`.
   That substring is the whole document.
3. Apply the fix-up below, then write the result to `harvest-api.json`.
4. Regenerate: `go run ./scripts/openapi/greenhouse/metadata`

Because there is no stable download URL, this depends on how the docs site is currently built. If
the extraction stops working, check whether ReadMe changed its embedding before assuming the spec
has moved.

## Required fix-up

The published spec is not valid OpenAPI in three places. Each declares an array's `items` as a bare
string instead of a schema object:

```
/v3/users/activate/bulk            → post.requestBody …properties.data.items
/v3/users/deactivate/bulk          → post.requestBody …properties.data.items
/v3/users/revoke_permissions/bulk  → post.requestBody …properties.data.items
```

Each is `"items": "number"` and must become `"items": {"type": "number"}`. Without this the
generator panics on load:

```
cannot unmarshal string into field Schema.items of type openapi3.Schema
```

All three are POST request bodies that the connector never reads, so the change affects only
whether the file parses.

## Excluded endpoints

`ignoreEndpoints` in the generator script documents what is deliberately left out and why — search
variants that duplicate their plain list endpoints without time-based filtering, and custom field
definition endpoints, which describe metadata rather than record data and are consumed internally
by `custom.go`.
