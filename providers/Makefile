# Generate Go types from OpenAPI spec.
.PHONY: gen
gen:
	oapi-codegen --config=oapi.config.yaml --o=types.gen.go --package=providers types.yaml

# Lint the OpenAPI spec.
.PHONY: lint
lint:
	rdme openapi:validate types.yaml
