# Metadata

The static file `objectsMetadata.json` is generated from the ClickUp OpenAPI spec in 
`scripts/openapi/clickup/internal/files/openapi.yaml` ([ClickUp OpenAPI spec](https://developer.clickup.com/docs/open-api-spec)).
The YAML is tracked with **Git LFS**.

Regenerate after updating the OpenAPI file:

```bash
go run ./scripts/openapi/clickup/metadata
```
