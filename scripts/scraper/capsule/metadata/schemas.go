package main

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type scrappedSchemas struct {
	Metadata *staticschema.Metadata[staticschema.FieldMetadataMapV2, any]
}

func newScrappedSchemas() scrappedSchemas {
	return scrappedSchemas{
		Metadata: staticschema.NewMetadata[staticschema.FieldMetadataMapV2](),
	}
}

func (s scrappedSchemas) SaveData(model scrapper.ModelDocLink, fieldName, fieldType, specialTag, description string) {
	modelDisplayName := displayNameMapping.Get(model.Name)
	fieldDisplayName := formatDisplay(fieldName)

	isReadOnly := strings.Contains(specialTag, "Read only")
	fieldValueOptions := filterOutExceptions(implyValueOptions(description))
	fieldType = formatFieldType(fieldType)

	var fields staticschema.FieldMetadataMapV2

	if strings.Contains(fieldName, Arrow) {
		// Special case. We do not care about nested fields.
		// The parent node should be saved as an object.
		// Example:
		// https://developer.capsulecrm.com/v2/models/task
		fieldName = strings.Split(fieldName, Arrow)[0]

		fields = staticschema.FieldMetadataMapV2{
			fieldName: staticschema.FieldMetadata{
				DisplayName:  formatDisplay(fieldName),
				ValueType:    common.ValueTypeOther,
				ProviderType: "object",
			},
		}
	} else {
		fields = staticschema.FieldMetadataMapV2{
			fieldName: staticschema.FieldMetadata{
				DisplayName:  fieldDisplayName,
				ValueType:    getFieldValueType(fieldType, fieldValueOptions),
				ProviderType: fieldType,
				ReadOnly:     goutils.Pointer(isReadOnly),
				Values:       getFieldValueOptions(fieldValueOptions),
			},
		}
	}

	responseKey := objectNameToResponseKey.Get(model.Name)
	urlPath := objectNameToURLPath.Get(model.Name)

	if model.Name == "projects" {
		// Duplicate projects object as kases.
		// https://developer.capsulecrm.com/v2/operations/Case
		s.Metadata.Add(
			common.ModuleRoot, "kases", modelDisplayName,
			urlPath, responseKey, fields, &model.URL, nil)
	}

	s.Metadata.Add(
		common.ModuleRoot, model.Name, modelDisplayName,
		urlPath, responseKey, fields, &model.URL, nil)
}

func filterOutExceptions(fieldValueOptions []string) []string {
	// Title is an exception which we shouldn't imply value options.
	// https://developer.capsulecrm.com/v2/models/party.
	// The title options should be coming from the CustomTitle object.
	// https://developer.capsulecrm.com/v2/operations/Custom_Title
	if len(fieldValueOptions) == 1 && fieldValueOptions[0] == "existing custom titles" {
		fieldValueOptions = nil
	}

	return fieldValueOptions
}
