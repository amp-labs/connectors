# Metadata

`objectsMetadata.json` is maintained manually from the [Dropbox API reference](https://docs.dropboxapi.com/dropbox-api/api-reference/all-endpoints).

The catalog uses the [objectcatalog](https://github.com/amp-labs/connectors/blob/main/internal/objectcatalog/README.md) format: each object has `displayName`, `path`, and `fields` (what `ListObjectMetadata` returns). Paths are the full RPC route under the API host, e.g. `/2/files/list_folder` for `POST https://api.dropboxapi.com/2/files/list_folder`.

The catalog base URL is `https://api.dropboxapi.com`. The **API version** (`/2`) is part of each object's `path`, not a separate module.

## `docs` on each object

Each object may include a `docs` URL pointing at the Dropbox endpoint used to define fields and read behavior. `ListObjectMetadata` ignores it (`common.ObjectMetadata` has no docs field), but keep `docs` in `objectsMetadata.json` — it documents where each object comes from when maintaining or extending the connector.

## Incremental read (Since / Until)

Dropbox list RPCs do not accept Since or Until. Read filters rows **after** each response using `timeField` in `supports.go` (e.g. `server_modified` for `files`, `created` for `file_requests`, `time_invited` for sharing lists). That matches datetime fields in this catalog. Lists are not sorted by those fields, so the connector still walks every page (`readhelper.Unordered`); filtering only changes which rows are returned, not how many API calls are made.
