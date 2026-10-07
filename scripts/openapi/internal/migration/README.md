# OpenAPI Extraction Migration Validation

## Description

When migrating an OpenAPI extraction script to use the new `scripts/openapi/internal/api` package, the generated output must be validated against the existing extraction results.

The migration validation script exists to ensure that switching to the new extraction package does not introduce any functional changes or degrade the metadata currently produced by the connector.

The old extraction scripts produce a `schemas.json` file, while the new extraction flow produces three artifacts:

* `objectsMetadata.json` — metadata describing the objects and their fields.
* `endpoints.json` — additional information about the API endpoints used to extract the objects.
* `queryParamsStats.json` — consolidated statistics about query parameters found across the extracted objects.

### Why compare the outputs?

The old `schemas.json` contains information from multiple modules, while the new extraction is module-agnostic. As a result, the new output may be split into multiple files and the information is organized differently.

`objectsMetadata.json` contains the metadata required to describe the objects and their fields. The `path` property is also preserved so developers can trace an object back to the endpoint from which its structure was inferred.

Other extraction-specific information, such as the HTTP method, response array location, and other endpoint details, is kept in `endpoints.json`.

The validation script accounts for these differences in representation and verifies that the new extraction still produces the same meaningful object metadata as the existing implementation.

The goal is simple: **the migration should change how the extraction is implemented, not what it produces.**

## Usage

Run the validation script by providing:

1. The existing `schemas.json`.
2. The new `objectsMetadata.json`.
3. The new `endpoints.json`.
4. The module name used by the existing schema.

For example:

```bash
go run scripts/openapi/internal/migration/validation.go \
  -old providers/sellsy/internal/metadata/schemas.json \
  -new providers/sellsy/internal/metadata/objectsMetadata.json.gz \
  -endpoints providers/sellsy/internal/metadata/endpoints.json \
  -module root
```

The command exits successfully when the old and new extraction results are equivalent. Any differences are reported as validation errors.

## Migration workflow

A typical migration should follow these steps:

1. Migrate the existing OpenAPI extraction script to use `scripts/openapi/internal/api`.
2. Generate the new `objectsMetadata.json` and `endpoints.json`.
3. Run the migration validation script.
4. Investigate and resolve any differences reported by the validator.
5. Once validation passes, remove the old extraction implementation.

The validation script is intended as a **migration tool**, not as part of the connector runtime.
