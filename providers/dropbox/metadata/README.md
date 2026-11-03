# Metadata

`schemas.json` is maintained manually from the [Dropbox API reference](https://docs.dropboxapi.com/dropbox-api/api-reference/all-endpoints).

The catalog base URL is `https://api.dropboxapi.com`. The **API version** (`/2`) is the module path on `root`, not part of each object path (same convention as stripping `/v1` on BambooHR or Notion). Example: RPC `POST /2/files/list_folder` is stored as module `/2` + object `/files/list_folder`.

## `docs` on each object

Each object may include a `docs` URL pointing at the Dropbox endpoint used to define fields and read behavior. `ListObjectMetadata` ignores it (`common.ObjectMetadata` has no docs field), but keep `docs` in `schemas.json` — it documents where each object comes from when maintaining or extending the connector.
