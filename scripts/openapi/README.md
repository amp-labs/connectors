# Description

This folder contains **scripts that process OpenAPI spec to produce Object Schemas**.
These schemas are later will be served via ListObjectMetadata.

# Structure

Scripts will be located under `scripts/openapi/<PROVIDER_NAME>/metadata/main.go`.

OpenAPI files that it loads can be found under `providers/<PROVIDER_NAME>/openapi/<FILE_NAME>.yaml|json`.
The output will be saved under `providers/<PROVIDER_NAME>/metadata/schemas.json`.

# Running instructions

Update OpenAPI file to the latest version, then execute the following command in the project root directory:

```
go run ./scripts/openapi/intercom/metadata
```
Check `providers/intercom/metadata/schemas.json` for any side effects. Please monitor the log output,
and if there are any errors, manually review the OpenAPI spec. Based on your review,
decide whether the endpoint should be integrated or ignored.

# Capability

These scripts offer:
* Control over which parts of the OpenAPI spec are relevant for processing.
* Formatting options for display names.
* Establishing relationship between Resource/Object Name and JSON field name, containing said object.
