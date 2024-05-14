package intercom

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

/*
Response example:

	{
	  "type": "list",
	  "data": [{...}],
	  "total_count": 1,
	  "pages": {
		"type": "pages",
		"page": 1,
		"next": "https://api.intercom.io/contacts/6643703ffae7834d1792fd30/notes?per_page=1&page=2",
		"per_page": 100,
		"total_pages": 1
	  }
	}

Note: `pages.next` can be null.
*/
func getTotalSize(node *ajson.Node) (int64, error) {
	return jsonquery.New(node).ArraySize("data")
}

func getRecords(node *ajson.Node) ([]map[string]any, error) {
	arr, err := jsonquery.New(node).Array("data", false)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(arr)
}

func makeNextRecordsURL(reqLink *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		next, err := getNextPageStringURL(node)
		if err == nil {
			return next, nil
		}

		if !errors.Is(err, jsonquery.ErrNotString) {
			// response from server doesn't meet any format that we expect
			return "", err
		}

		// Probably, we are dealing with an object under `pages.next`
		startingAfter, err := jsonquery.New(node, "pages", "next").Str("starting_after", true)
		if err != nil {
			return "", err
		}

		if startingAfter == nil {
			// next page doesn't exist
			return "", nil
		}

		reqLink.WithQueryParam("starting_after", *startingAfter)

		return reqLink.String(), nil
	}
}

// Some responses have full URL stored at `pages.next`.
func getNextPageStringURL(node *ajson.Node) (string, error) {
	nextPage, err := jsonquery.New(node, "pages").Str("next", true)
	if err != nil {
		return "", err
	}

	if nextPage == nil {
		return "", nil
	}

	return *nextPage, nil
}

func getMarshaledData(records []map[string]interface{}, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	for i, record := range records {
		data[i] = common.ReadResultRow{
			Fields: common.ExtractLowercaseFieldsFromRaw(fields, record),
			Raw:    record,
		}
	}

	return data, nil
}

// func getList(node *ajson.Node) (*ajson.Node, error) {
//	jsonquery.New(node)
//}
