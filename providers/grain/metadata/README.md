# Metadata

Grain does not provide an OpenAPI specification or a schema discovery endpoint.
The static file `schemas.json` was created manually from the "Common Response Objects"
section of the [Grain API documentation](https://developers.grain.com/endpoints.html)
(API version `2026-10-01`) and is embedded by `metadata.go`.

## Supported Objects

| Object          | Endpoint                  | Response key    |
|-----------------|---------------------------|-----------------|
| `recordings`    | `POST /v2/recordings`     | `recordings`    |
| `uploads`       | `POST /v2/uploads`        | `uploads`       |
| `hooks`         | `POST /v2/hooks`          | `hooks`         |
| `users`         | `POST /v2/users`          | `users`         |
| `teams`         | `POST /v2/teams`          | `teams`         |
| `meeting_types` | `POST /v2/meeting_types`  | `meeting_types` |

Enumerated values (`media_type`, `source`, `share_state`, upload `status`, `hook_type`
and meeting type `scope`) are taken from the documentation and exposed as `singleSelect`.

## Live-observed fields

The following fields are returned by the live API (observed 2026-10-28) but are not
declared in the documentation:

- `recordings.sensitivity`
- `recordings.sensitivity_source`

## Include-gated fields

Some `recordings` and `uploads` fields are only returned when requested through the
`include` request parameter (for example `participants`, `ai_summary`, `attributes`).
