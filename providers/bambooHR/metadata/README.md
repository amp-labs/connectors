# Metadata

The static file `schemas.json` is generated from `openapi/openapi.json` (sourced from the [BambooHR OpenAPI spec](https://openapi.bamboohr.io/main/latest/docs/openapi/public-openapi.yaml); see [Postman collection docs](https://documentation.bamboohr.com/docs/postman-collection)).

Regenerate after updating `openapi/openapi.json`:

```bash
go run ./scripts/openapi/bambooHR/metadata
```

Object paths in `schemas.json` are relative to the connector base URL (`https://{workspace}.bamboohr.com/api`), so OpenAPI `/api/v1/...` routes are stored as `/v1/...`.
