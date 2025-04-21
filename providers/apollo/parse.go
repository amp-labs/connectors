package apollo

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsQuery returns the URL for the next page of results.
func getNextRecords(node *ajson.Node) (string, error) {
	var nextPage string

	pagination, err := jsonquery.New(node).ObjectOptional("pagination")
	if err != nil {
		return "", err
	}

	if pagination != nil {
		page, err := jsonquery.New(pagination).IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		totalPages, err := jsonquery.New(pagination).IntegerOptional("total_pages")
		if err != nil {
			return "", err
		}

		if page < *totalPages {
			nextPage = strconv.FormatInt(page+1, 10)
		}
	}

	return nextPage, nil
}

// recordsWrapperFunc returns the records using the objectName dynamically.
// It handles both root-level arrays and nested arrays under an object name.
func recordsWrapperFunc(obj string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		// If obj is empty or node is an array, process the node directly
		if obj == "" || node.IsArray() {
			children, err := node.GetArray()
			if err != nil {
				return nil, err
			}
			return jsonquery.Convertor.ArrayToMap(children)
		}

		// If not a root array, try to get array from the specified object field
		result, err := jsonquery.New(node).ArrayOptional(obj)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(result)
	}
}

// searchRecords returns a function that parses the search requests response.
func searchRecords(fld string) common.RecordsFunc {
	var records []map[string]any

	fld = constructSupportedObjectName(fld)

	return func(node *ajson.Node) ([]map[string]any, error) {
		result, err := jsonquery.New(node).ArrayOptional(fld)
		if err != nil {
			return nil, err
		}

		rec, err := jsonquery.Convertor.ArrayToMap(result)
		if err != nil {
			return nil, err
		}

		records = append(records, rec...)

		return records, nil
	}
}
