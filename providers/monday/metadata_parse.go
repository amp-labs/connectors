package monday

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func introspectionQueryForObject(objectName string) string {
	typeName := naming.NewSingularString(naming.CapitalizeFirstLetterEveryWord(objectName)).String()

	return fmt.Sprintf(`{
		__type(name: "%s") {
			name
			fields {
				name
				type {
					name
					kind
					ofType {
						name
					}
				}
			}
		}
	}`, typeName)
}

func parseSingleObjectMetadataResponse(
	objectName string,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(common.FieldsMetadata),
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	metadataResp, err := common.UnmarshalJSON[MetadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(metadataResp.Data.Type.Fields) == 0 {
		return nil, fmt.Errorf(
			"missing or empty fields for object: %s, error: %w",
			objectName,
			common.ErrMissingExpectedValues,
		)
	}

	for _, field := range metadataResp.Data.Type.Fields {
		objectMetadata.AddFieldMetadata(field.Name, common.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    common.ValueTypeOther,
			ProviderType: "",
		})
	}

	return &objectMetadata, nil
}
