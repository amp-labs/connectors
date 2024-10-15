package pipedrive

import (
	"encoding/json"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// nextRecordsURL builds the next-page url func.
func nextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// check if there is more items in the collection.
		more, err := jsonquery.New(node, "additional_data", "pagination").Bool("more_items_in_collection", true)
		if err != nil {
			return "", err
		}

		startValue, err := jsonquery.New(node, "additional_data", "pagination").Integer("next_start", true)
		if err != nil {
			return "", err
		}

		if *more {
			url.WithQueryParam("start", strconv.FormatInt(*startValue, 10))

			return url.String(), nil
		}

		return "", nil
	}
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	var d responseData

	b := node.Source()
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}

	records := constructRecords(d)

	return records, nil
}

func constructRecords(d responseData) []map[string]any {
	records := make([]map[string]any, len(d.Data))

	for i, record := range d.Data {
		recordItems := make(map[string]any)

		for k, v := range record {
			recordItems[k] = v
		}

		records[i] = recordItems
	}

	return records
}
