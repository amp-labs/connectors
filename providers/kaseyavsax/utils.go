package kaseyavsax

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func inferValueTypeFromData(value any) common.ValueType {
	v := reflect.ValueOf(value)

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

func records() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node).ArrayOptional(dataField)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		meta := jsonquery.New(node, metaField)

		nextPage, err := meta.StringOptional(NextRecordsField)
		if err != nil {
			return "", err
		}

		if nextPage == nil {
			return "", nil
		}

		return *nextPage, nil
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}
	writeSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}

// https://api.vsax.net/#get-all-workflows
// https://api.vsax.net/#get-all-tasks
// https://api.vsax.net/#get-all-scopes
var objectsWithUpdateAtFields = datautils.NewStringSet("automation/workflows", //nolint:gochecknoglobals
	"automation/tasks", "scopes")

func supportsFiltering(objectName string) bool {
	return objectsWithUpdateAtFields.Has(objectName)
}
