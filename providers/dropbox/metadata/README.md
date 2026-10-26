# Metadata

`schemas.json` is maintained manually from the [Dropbox API reference](https://docs.dropboxapi.com/dropbox-api/api-reference/all-endpoints). Dropbox exposes an [OpenAPI specification](https://docs.dropboxapi.com/dropbox-api/llms.txt), but the published download URL currently returns HTML rather than the spec file, so fields were taken from the endpoint docs (union types merged where one list returns multiple shapes).

The catalog base URL is `https://api.dropboxapi.com`. Object paths are the RPC route (for example `POST /2/files/list_folder` is stored as `/2/files/list_folder`).
