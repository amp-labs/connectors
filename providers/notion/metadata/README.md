# Metadata

`schemas.json` is generated from the Notion OpenAPI spec at `scripts/openapi/notion/internal/files/openapi.json` ([Notion API reference](https://developers.notion.com/reference/intro)). The spec is stored with Git LFS. After clone, run `git lfs pull` before regenerating schemas.

```bash
go run ./scripts/openapi/notion/metadata
```

The catalog base URL is `https://api.notion.com`. Each object path is the resource only, without the `v1` version segment (`/v1/blocks` becomes `/blocks`).
