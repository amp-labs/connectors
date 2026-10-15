package graphql

import (
	_ "embed"
)

// SchemaFile is the Jump Enterprise GraphQL schema (SDL).
// Jump disables introspection, so the schema is downloaded with the API key instead.
// https://my.jumpapp.com/enterprise/documentation (API Reference)
//
//go:embed schema.graphql
var SchemaFile []byte //nolint:gochecknoglobals
