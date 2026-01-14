# Metadata

The static file `schemas.json` is generated from the CloudTalk OpenAPI specification (`swagger.json`).

## Generation

To regenerate the schemas, run the following command from the root of the repository:

```bash
go run scripts/openapi/cloudtalk/metadata/main.go
```

## OpenAPI Spec

The OpenAPI specification file is located at `providers/cloudtalk/metadata/openapi/swagger.json`. It was downloaded from [CloudTalk Developers](https://developers.cloudtalk.io/swagger.json) and patched to fix structural issues (invalid references, incorrect types) to allow for successful parsing and schema generation.
