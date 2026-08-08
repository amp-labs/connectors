# Metadata

`schemas.json` is generated from Monaco's public OpenAPI spec. The spec needs no
authentication, so it can be refreshed at any time:

```sh
curl -s -o providers/monaco/metadata/openapi/openapi.json https://api.monaco.com/openapi.json
go run ./scripts/openapi/monaco/metadata
```

## Notes

**Two read shapes.** The eight core collections are listed over `POST /v1/<object>/list`
with a JSON body (`page`, `page_size`, `filters`, `sort`) and respond with
`{data, pagination, meta}`. `tags`, `users` and `sequenceTemplates` are listed over
plain `GET` and respond with `{data, meta}` — no `pagination` object, no next page.
The generator reads both operations and combines them; records are always under `data`.

**Paths are split across module and object.** `staticschema.Add` hoists the longest
common prefix into the module path, so the module holds `/v1` and each object holds
only the remainder (`/contacts/list`). Join the two to rebuild a request URL.

**Nullable fields are collapsed.** Monaco marks nearly every response field
`anyOf: [{type: X}, {type: null}]`. A schema whose type is buried in an `anyOf` has no
`type` of its own, so the extractor would classify it as `other` — which left ~62 of
125 fields untyped, including obvious strings like `contacts.email`. `openapi/file.go`
rewrites the unambiguous single-non-null case before the explorer sees it, bringing
that down to 17 fields that are genuinely arrays or nested objects. The vendored
`openapi.json` is left byte-for-byte as downloaded.

## Not covered here

Monaco serves `GET /v1/schemas/{entity}` at runtime, which reports per-org **custom
fields** (`custom_field_<uuid>`) and each field's `allowed_operators`. It covers only
6 of the 11 objects (accounts, contacts, meetings, opportunities, sequences, tasks),
so it cannot replace the static schemas. Layering it in as a composite provider is
deferred until we have credentials to verify the payload shape. The `allowed_operators`
data will also matter when the search action lands, since it defines which
`filters[].condition` values each field accepts.
