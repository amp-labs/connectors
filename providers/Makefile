# Generate Go types from OpenAPI spec.

CATALOG_YAML = https://raw.githubusercontent.com/amp-labs/openapi/089a1627afc8bcdabd8e7ec37f142e0992045ff5/catalog/catalog.yaml


.PHONY: gen/catalog
gen/catalog:
	oapi-codegen --config=oapi.config.yaml --o=types.gen.go $(CATALOG_YAML)
