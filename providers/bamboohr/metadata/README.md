# Metadata

The static file `schemas.json` is generated from `openapi/openapi.yaml` ([BambooHR OpenAPI spec](https://openapi.bamboohr.io/main/latest/docs/openapi/public-openapi.yaml); see [Postman collection docs](https://documentation.bamboohr.com/docs/postman-collection)). The YAML is stored with **Git LFS** (in git you only see a small pointer file, like Recurly). After clone, run `git lfs pull` before building or regenerating schemas.

Regenerate after updating `openapi/openapi.yaml` or manual field overrides in `scripts/openapi/bamboohr/metadata/manual-objects.json`:

```bash
go run ./scripts/openapi/bamboohr/metadata
```

The catalog base URL is `https://{workspace}.bamboohr.com/api/v1`. Each object `path` in `schemas.json` is the remainder after that prefix (for example OpenAPI `/api/v1/employees/directory` → object path `/employees/directory`). The API version is not repeated on individual objects.
