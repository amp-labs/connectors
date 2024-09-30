// nolint
package attio

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

var dummyNextPageFunc = func(*ajson.Node) (string, error) {
	return "", nil
}

// recordsWrapperFunc returns the records using the objectName dynamically.
func recordsWrapperFunc(obj string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		result, err := jsonquery.New(node).Array(obj, true)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(result)
	}
}
