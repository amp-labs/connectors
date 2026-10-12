package freshdesk

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// extractRecords reads the records held at the root of a response body.
// Collections are served as a list, while a singleton such as settings answers with a lone
// object, which stands for a single record. Common only extracts lists, hence the local reader.
func extractRecords(node *ajson.Node) ([]map[string]any, error) {
	if !node.IsArray() {
		record, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		return []map[string]any{record}, nil
	}

	array, err := node.GetArray()
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(array)
}

// nextRecordsURL reads the page that follows from the Link header.
// Freshdesk sends it only while one remains, so its absence means the read is done.
// Ex: `<https://ampersand.freshdesk.com/api/v2/contacts?per_page=1&page=2>; rel="next"`.
func nextRecordsURL(response *common.JSONHTTPResponse) common.NextPageFunc {
	return func(*ajson.Node) (string, error) {
		return httpkit.HeaderLink(response, "next"), nil
	}
}
