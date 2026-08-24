package slack

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// validateResponse checks the Slack-specific "ok" field (Slack always returns HTTP 200,
// even on failure), interprets any error code, and returns the records array for the given object.
func validateResponse(node *ajson.Node) error {
	// Slack always returns HTTP 200, even when the request fails. The "ok" field
	// in the response body is the real indicator of success or failure.
	ok, err := jsonquery.New(node).BoolRequired("ok")
	if err != nil {
		return err
	}

	if !ok {
		// Map the Slack error code to a sentinel error so callers can use
		// errors.Is to react appropriately (re-auth, retry, etc.).
		errorCode, err := jsonquery.New(node).StringOptional("error")
		if err != nil {
			return err
		}

		if errorCode != nil {
			return interpretSlackErrorCode(*errorCode)
		}

		return common.ErrBadProviderResponse
	}

	return nil
}

// getResponseCollectionRecords parses node to get a list of records.
func getResponseCollectionRecords(node *ajson.Node, responseField string) ([]*ajson.Node, error) {
	if err := validateResponse(node); err != nil {
		return nil, err
	}

	return jsonquery.New(node).ArrayRequired(responseField)
}

// getResponseSingleRecord parses node to get a single record.
// Example response:
//
//	{
//	   "ok": true,					-> used by validateResponse
//	   "channel": {					-> unwrapped by ObjectRequired
//	       "id": "C0B9RLDTM35",
//	       "created": 1781192018,
//	       "creator": "U0B9R8LFG1F",
//		  }
//	}
func getResponseSingleRecord(node *ajson.Node, responseField string) (*ajson.Node, error) {
	if err := validateResponse(node); err != nil {
		return nil, err
	}

	return jsonquery.New(node).ObjectRequired(responseField)
}

func recordsFunc(responseField string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		arr, err := getResponseCollectionRecords(node, responseField)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

func nodeRecords(responseField string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return getResponseCollectionRecords(node, responseField)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return jsonquery.New(node, "response_metadata").StrWithDefault("next_cursor", "")
	}
}
