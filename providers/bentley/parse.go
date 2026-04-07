package bentley

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/bentley/metadata"
	"github.com/spyzhov/ajson"
)

func getRecords(moduleID common.ModuleID, objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		responseFieldName := metadata.Schemas.LookupArrayFieldName(moduleID, objectName)

		arr, err := jsonquery.New(node).ArrayOptional(responseFieldName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		href, err := jsonquery.New(node, "_links", "next").StrWithDefault("href", "")
		// If there's an error (e.g. _links or next doesn't exist), or if href is empty, return "" with no error,
		// This is expected because some of the objects don't have pagination,
		// so the absence of _links.next.href just means there are no more pages.
		if err != nil {
			return "", nil //nolint:nilerr
		}

		if len(href) == 0 {
			return "", nil
		}

		return href, nil
	}
}
