package slack

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getSlackResponseRecords checks the Slack-specific "ok" field (Slack always returns HTTP 200,
// even on failure), interprets any error code, and returns the records array for the given object.
func getSlackResponseRecords(node *ajson.Node, objectName string) ([]*ajson.Node, error) {
	// Slack always returns HTTP 200, even when the request fails. The "ok" field
	// in the response body is the real indicator of success or failure.
	ok, err := jsonquery.New(node).BoolRequired("ok")
	if err != nil {
		return nil, err
	}

	if !ok {
		// Map the Slack error code to a sentinel error so callers can use
		// errors.Is to react appropriately (re-auth, retry, etc.).
		errorCode, err := jsonquery.New(node).StringOptional("error")
		if err != nil {
			return nil, err
		}

		if errorCode != nil {
			return nil, interpretSlackErrorCode(*errorCode)
		}

		return nil, common.ErrBadProviderResponse
	}

	return jsonquery.New(node).ArrayRequired(objectResponseField.Get(objectName))
}

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		arr, err := getSlackResponseRecords(node, objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

func nodeRecords(objectName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return getSlackResponseRecords(node, objectName)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node, "response_metadata").StrWithDefault("next_cursor", "")
	}
}
