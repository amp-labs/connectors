# Metadata

`objectsMetadata.json` and `endpoints.json` are generated from the Notion OpenAPI spec at `scripts/openapi/notion/internal/files/openapi.json` ([Notion API reference](https://developers.notion.com/reference/intro)). The spec is stored with Git LFS. After clone, run `git lfs pull` before regenerating.

```bash
go run ./scripts/openapi/notion/metadata
```

The catalog base URL is `https://api.notion.com`. Each object path is the resource only, without the `v1` version segment (`/v1/users` becomes `/users`).

Supported objects in the generated catalog are parent-free `GET` lists that Read implements today: `users` and `file_uploads`. The generator reads those paths from the OpenAPI spec. Both objects are also loaded from component schemas (`userObjectResponse`, `fileUploadObjectResponse`) because the list extractor leaves some `oneOf` fields without a type and does not collect `const` values into enums. See `walkSchema` in `scripts/openapi/notion/metadata/main.go`.

Related list endpoints not in the catalog yet:

- **custom_emojis** — workspace list with the same `results` pagination as users; no parent id ([list custom emojis](https://developers.notion.com/reference/list-custom-emojis)).
- **views** — GET list but requires `database_id` or `data_source_id` ([list views](https://developers.notion.com/reference/list-views)).

These are not Read objects today: pages and data sources (discoverable only via `POST /v1/search`, no time window), blocks and comments (list requires a parent id), databases (no workspace-wide list in the public API), search and agent query endpoints, OAuth, and webhooks. Subscribe may cover databases, data sources, and pages when Notion's webhook events align with customer needs, even though Read cannot enumerate them incrementally.
