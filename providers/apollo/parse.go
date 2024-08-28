package apollo

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsQuery returns the URL for the next page of results.
func getNextRecords(node *ajson.Node) (string, error) {
	var nextPage string

	pagination, err := jsonquery.New(node).Object("pagination", true)
	if err != nil {
		return "", err
	}

	if pagination != nil {
		page, err := jsonquery.New(pagination).IntegerWithDefault("page", 1)
		if err != nil {
			return "", err
		}

		totalPages, err := jsonquery.New(pagination).Integer("total_pages", true)
		if err != nil {
			return "", err
		}

		if page < *totalPages {
			nextPage = fmt.Sprint(page + 1)
		}
	}

	return nextPage, nil
}

// rcordsWrapperFunc returns the records using the objectName dynamically.
func recordsWrapperFunc(obj string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		result, err := jsonquery.New(node).Array(obj, true)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(result)
	}
}

// rcordsWrapperFunc returns the records using the objectName dynamically.
func recordsSizeWrapperFunc(obj string) common.ListSizeFunc {
	return func(node *ajson.Node) (int64, error) {
		result, err := jsonquery.New(node).Array(obj, true)
		if err != nil {
			return 0, err
		}

		rcds, err := jsonquery.Convertor.ArrayToMap(result)
		if err != nil {
			return 0, err
		}

		size := int64(len(rcds))

		return size, nil
	}
}
