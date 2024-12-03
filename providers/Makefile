# Generate Go types from OpenAPI spec.

CATALOG_YAML = https://raw.githubusercontent.com/amp-labs/openapi/main/catalog/catalog.yaml


.PHONY: gen/catalog
gen/catalog:
	oapi-codegen --config=oapi.config.yaml --o=types.gen.go $(CATALOG_YAML)
