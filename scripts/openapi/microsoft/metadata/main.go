package main

import (
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/microsoft/metadata"
)

func main() {
	definitions, err := extractSchemas()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()

	for _, object := range definitions {
		for _, field := range object.Fields {
			// Schema naming is usually in singular case.
			// At the same time the URL endpoints are plural.
			// Objects, therefore, will be plural.
			schemas.Add(common.ModuleRoot, object.ObjectName, object.DisplayName, "", "",
				fieldToMetadata(field), nil, "")
		}
	}

	goutils.MustBeNil(metadata.FileManager.FlushSchemas(schemas))

	slog.Info("Completed.")
}

func fieldToMetadata(field metadatadef.Field) staticschema.FieldMetadataMapV2 {
	return staticschema.FieldMetadataMapV2{
		field.Name: staticschema.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    getFieldValueType(field),
			ProviderType: field.Type,
		},
	}
}

// https://docs.oasis-open.org/odata/odata/v4.0/csprd02/part3-csdl/odata-v4.0-csprd02-part3-csdl.html#_Toc360208768
func getFieldValueType(field metadatadef.Field) common.ValueType { // nolint:cyclop
	switch field.Type {
	case "Edm.Binary":
		return common.ValueTypeString
	case "Edm.Boolean":
		return common.ValueTypeBoolean
	case "Edm.Date":
		return common.ValueTypeDate
	case "Edm.DateTimeOffset":
		return common.ValueTypeDateTime
	case "Edm.Double":
		return common.ValueTypeFloat
	case "Edm.Duration":
		return common.ValueTypeDateTime
	case "Edm.Guid":
		return common.ValueTypeString
	case "Edm.Int32":
		return common.ValueTypeInt
	case "Edm.Int64":
		return common.ValueTypeInt
	case "Edm.Single":
		return common.ValueTypeString
	case "Edm.Stream":
		return common.ValueTypeOther
	case "Edm.String":
		return common.ValueTypeString
	case "Edm.TimeOfDay":
		return common.ValueTypeDateTime
	default:
		return common.ValueTypeOther
	}
}
