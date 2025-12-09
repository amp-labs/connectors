package main

import (
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/justcall/metadata"
	"github.com/getkin/kin-openapi/openapi3"
)

const openAPIFilePath = "providers/justcall/openapi/api-v21.json"

func loadOpenAPIDocument() *openapi3.T {
	data, err := os.ReadFile(openAPIFilePath)
	goutils.MustBeNil(err)

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData(data)
	goutils.MustBeNil(err)

	return doc
}

// objectConfig defines the mapping from API path to object configuration.
type objectConfig struct {
	objectName  string
	displayName string
}

// nolint:gochecknoglobals
var objectEndpoints = map[string]objectConfig{
	"/v2.1/calls":                       {objectName: "calls", displayName: "Calls"},
	"/v2.1/calls_ai":                    {objectName: "calls_ai", displayName: "Calls AI"},
	"/v2.1/contacts":                    {objectName: "contacts", displayName: "Contacts"},
	"/v2.1/contacts/blacklist":          {objectName: "blacklisted-contacts", displayName: "Blacklisted Contacts"},
	"/v2.1/meetings_ai":                 {objectName: "meetings_ai", displayName: "Meetings AI"},
	"/v2.1/phone-numbers":               {objectName: "phone-numbers", displayName: "Phone Numbers"},
	"/v2.1/sales_dialer/calls":          {objectName: "sales_dialer/calls", displayName: "Sales Dialer Calls"},
	"/v2.1/sales_dialer/campaigns":      {objectName: "sales_dialer/campaigns", displayName: "Sales Dialer Campaigns"},
	"/v2.1/sales_dialer/contacts":       {objectName: "sales_dialer/contacts", displayName: "Sales Dialer Contacts"},
	"/v2.1/texts":                       {objectName: "texts", displayName: "Texts"},
	"/v2.1/texts/tags":                  {objectName: "texts/tags", displayName: "SMS Tags"},
	"/v2.1/user_groups":                 {objectName: "user_groups", displayName: "User Groups"},
	"/v2.1/users":                       {objectName: "users", displayName: "Users"},
	"/v2.1/webhooks":                    {objectName: "webhooks", displayName: "Webhooks"},
	"/v2.1/whatsapp/messages":           {objectName: "whatsapp/messages", displayName: "WhatsApp Messages"},
	"/v2.1/whatsapp/messages/templates": {objectName: "whatsapp/templates", displayName: "WhatsApp Templates"},
}

func main() {
	// Load OpenAPI document
	doc := loadOpenAPIDocument()

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()

	// Process each configured endpoint
	for path, config := range objectEndpoints {
		pathItem := doc.Paths.Find(path)
		if pathItem == nil {
			slog.Warn("path not found in OpenAPI", "path", path)

			continue
		}

		// Get the GET operation
		getOp := pathItem.Get
		if getOp == nil {
			slog.Warn("GET operation not found", "path", path)

			continue
		}

		// Get 200 response schema
		resp200 := getOp.Responses.Status(200)
		if resp200 == nil || resp200.Value == nil {
			slog.Warn("200 response not found", "path", path)

			continue
		}

		content := resp200.Value.Content.Get("application/json")
		if content == nil || content.Schema == nil {
			slog.Warn("JSON schema not found", "path", path)

			continue
		}

		schema := content.Schema.Value
		if schema == nil {
			slog.Warn("schema value is nil", "path", path)

			continue
		}

		// JustCall OpenAPI has inconsistent schemas:
		// - Some describe item properties directly (calls, users, etc.)
		// - Some describe envelope with data array (contacts/blacklist, etc.)
		// Check if this is an envelope schema by looking for "data" array property.
		itemSchema := schema
		if dataSchema, hasData := schema.Properties["data"]; hasData {
			// This is an envelope schema - get the items schema from data array
			if dataSchema.Value != nil && dataSchema.Value.Items != nil && dataSchema.Value.Items.Value != nil {
				itemSchema = dataSchema.Value.Items.Value
			}
		}

		// Extract fields from the item schema properties.
		for propName, propSchema := range itemSchema.Properties {
			fieldMeta := extractFieldMetadata(propName, propSchema)
			schemas.Add(common.ModuleRoot, config.objectName, config.displayName, path, "data",
				fieldMeta, nil, false)
		}

		slog.Info("extracted object",
			"objectName", config.objectName,
			"fields", len(itemSchema.Properties),
		)
	}

	goutils.MustBeNil(metadata.FileManager.SaveSchemas(schemas))

	slog.Info("Completed.")
}

// extractFieldMetadata converts an OpenAPI schema property to FieldMetadataMapV2.
func extractFieldMetadata(name string, schemaRef *openapi3.SchemaRef) staticschema.FieldMetadataMapV2 {
	schema := schemaRef.Value
	if schema == nil {
		return staticschema.FieldMetadataMapV2{
			name: staticschema.FieldMetadata{
				DisplayName:  name,
				ValueType:    common.ValueTypeOther,
				ProviderType: "unknown",
			},
		}
	}

	providerType := getSchemaType(schema)
	valueType := getValueType(providerType)

	// Check for enum values
	var values staticschema.FieldValues
	if len(schema.Enum) > 0 {
		valueType = common.ValueTypeSingleSelect
		values = make(staticschema.FieldValues, len(schema.Enum))

		for i, v := range schema.Enum {
			strVal, ok := v.(string)
			if !ok {
				strVal = ""
			}

			values[i] = staticschema.FieldValue{
				Value:        strVal,
				DisplayValue: strVal,
			}
		}
	}

	return staticschema.FieldMetadataMapV2{
		name: staticschema.FieldMetadata{
			DisplayName:  name,
			ValueType:    valueType,
			ProviderType: providerType,
			Values:       values,
		},
	}
}

// getSchemaType extracts the type string from OpenAPI schema.
func getSchemaType(schema *openapi3.Schema) string {
	if schema.Type != nil && len(*schema.Type) > 0 {
		return (*schema.Type)[0]
	}

	return "unknown"
}

func getValueType(providerType string) common.ValueType {
	switch providerType {
	case "integer", "number":
		return common.ValueTypeInt
	case "boolean":
		return common.ValueTypeBoolean
	case "string":
		return common.ValueTypeString
	case "array", "object":
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}
