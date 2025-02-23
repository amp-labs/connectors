package helpscout

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL(root *ajson.Node) (string, error) {
	next, err := jsonquery.New(root, "_links").ObjectOptional("next")
	if err != nil {
		return "", err
	}

	if next == nil {
		return "", nil
	}

	url, err := jsonquery.New(next).StringOptional("href")
	if err != nil {
		return "", err
	}

	return *url, nil
}

func getRecords(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		rcds, err := jsonquery.New(node, "_embedded").ArrayOptional(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(rcds)
	}
}
