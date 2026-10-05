# Metadata

The static file `schemas.json` is generated from the BambooHR OpenAPI spec in `scripts/openapi/bamboohr/internal/files/openapi.yaml` ([BambooHR OpenAPI spec](https://openapi.bamboohr.io/main/latest/docs/openapi/public-openapi.yaml)). The YAML is stored with **Git LFS**. After clone, run `git lfs pull` before building or regenerating schemas.

Regenerate after updating the OpenAPI file or manual field overrides in `scripts/openapi/bamboohr/internal/files/manual-objects.json`:

```bash
go run ./scripts/openapi/bamboohr/metadata
```

The catalog base URL is `https://{workspace}.bamboohr.com/api/v1`. Each object `path` in `schemas.json` is the remainder after that prefix (for example OpenAPI `/api/v1/employees/directory` → object path `/employees/directory`). The API version is not repeated on individual objects; the connector adds `v1` when building request URLs.
