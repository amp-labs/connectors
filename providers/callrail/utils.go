package callrail

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var responseField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint: gochecknoglobals
	"integration_triggers": "integration_criteria",
	"text-messages":        "conversations",
}, func(objectname string) string {
	return objectname
})

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		recordField := responseField.Get(objectName)

		records, err := jsonquery.New(node).ArrayOptional(recordField)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		var nextPage string

		page, err := jsonquery.New(node).IntegerOptional("page")
		if err != nil {
			return "", err
		}

		totalPage, err := jsonquery.New(node).IntegerOptional("total_pages")
		if err != nil {
			return "", err
		}

		if totalPage == nil || page == nil {
			return "", nil
		}

		if *page < *totalPage {
			nextPage = strconv.Itoa(int(*page + 1))
		}

		return nextPage, nil
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
