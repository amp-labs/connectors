package marketo

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextPageToken", "")
}

func constructNextRecordsURL(object string) common.NextPageFunc {
	if filtersByIDs(object) {
		// We use a new func for generating the next page url.
		return func(node *ajson.Node) (string, error) {
			jsonParser := jsonquery.New(node)

			data, err := jsonParser.Array("result", false)
			if err != nil {
				return "", err
			}

			// If the records returned matches the maximum batchsize, there is a high probability of having more records.
			// We'd have to check for the next page records, also due deletes the is also a probability of having more records
			// even if the size do not reach 300.
			if len(data) > 0 {
				id, err := jsonquery.New(data[len(data)-1]).Integer("id", false)
				if err != nil {
					return "", err
				}

				return strconv.Itoa(int(*id) + 1), nil
			}

			return "", nil
		}
	}

	return getNextRecordsURL
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
