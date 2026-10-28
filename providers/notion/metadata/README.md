# Metadata

`schemas.json` is generated from the Notion OpenAPI spec at `scripts/openapi/notion/internal/files/openapi.json` ([Notion API reference](https://developers.notion.com/reference/intro)). The spec is stored with Git LFS. After clone, run `git lfs pull` before regenerating schemas.

```bash
go run ./scripts/openapi/notion/metadata
```

The catalog base URL is `https://api.notion.com`. Each object path is the resource only, without the `v1` version segment (`/v1/users` becomes `/users`).

Supported objects in `schemas.json` are parent-free `GET` lists: `users` and `file_uploads`. The generator reads those paths from the OpenAPI spec. Both objects are also loaded from component schemas (`userObjectResponse`, `fileUploadObjectResponse`) because the list extractor leaves some `oneOf` fields without a type and does not collect `const` values into enums. See `walkSchema` in `scripts/openapi/notion/metadata/main.go`.

These are not connector objects: pages and data sources (listed only via `POST /v1/search`, no time window), blocks and comments and views (list requires a parent id), databases (create-only path), search and agent query endpoints, OAuth, and webhooks.
