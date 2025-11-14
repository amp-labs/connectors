package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	//go:embed samples.json
	samplesData []byte

	outputCopper = scrapper.NewWriter[staticschema.FieldMetadataMapV2]( // nolint:gochecknoglobals
		fileconv.NewPath("providers/copper/internal/metadata"))
)

// Script uses object samples from Documentation https://developer.copper.com/index.html.
// Each sample is used to extract field names to construct schema.json which will be returned via ListObjectMetadata.
func main() {
	var definitions []objectDefinition

	err := json.Unmarshal(samplesData, &definitions)
	goutils.MustBeNil(err)

	schemas := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()

	for _, definition := range definitions {
		for _, field := range definition.GetFields() {
			url := strings.ReplaceAll(definition.URL, "https://api.copper.com/developer_api/v1", "")
			schemas.Add(common.ModuleRoot, definition.Name, definition.DisplayName, url,
				definition.ResponseKey, field, nil, nil)
		}
	}

	goutils.MustBeNil(outputCopper.FlushSchemas(schemas))

	log.Println("Completed.")
}

type objectDefinition struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"displayName"`
	ResponseKey string         `json:"responseKey"`
	URL         string         `json:"url"`
	Sample      map[string]any `json:"sample"`
}

func (d objectDefinition) GetFields() []staticschema.FieldMetadataMapV2 {
	fields := make([]staticschema.FieldMetadataMapV2, 0)

	for fieldName, value := range d.Sample {
		primitiveType := getPrimitiveType(value)
		fields = append(fields, staticschema.FieldMetadataMapV2{
			fieldName: staticschema.FieldMetadata{
				DisplayName:  fieldName,
				ValueType:    primitiveType,
				ProviderType: string(primitiveType),
				Values:       nil,
			},
		})
	}

	return fields
}

func getPrimitiveType(value any) common.ValueType {
	switch value.(type) {
	case string:
		return common.ValueTypeString
	case bool:
		return common.ValueTypeBoolean
	case float32, float64:
		return common.ValueTypeFloat
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return common.ValueTypeInt
	default:
		return common.ValueTypeOther
	}
}
