package marketo

import (
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextPageToken", "")
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	result, err := jsonquery.New(node).Array("result", true)
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(result)
}

func usesStandardId(object string) bool {
	for _, v := range IdResponseObjects {
		if v == object {
			return true
		}
	}

	return false
}

func usesMarketoGUID(object string) bool {
	for _, v := range marketoGUIDResponseObjects {
		if v == object {
			return true
		}
	}

	return false
}
