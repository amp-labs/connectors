package merging

import "github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"

func CombineByObjectName(schema spec.Schema) string {
	// Use objectName as ID
	return schema.ObjectName
}

func ChooseRight(left spec.Schema, right spec.Schema) spec.Schema {
	return right
}

func ChooseLeft(left spec.Schema, right spec.Schema) spec.Schema {
	return left
}
