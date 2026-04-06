package slack

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var errResponseIndicatesFailure = errors.New("response indicated failure")

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
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

			return nil, errResponseIndicatesFailure
		}

		responseKey := objectResponseField.Get(objectName)

		arr, err := jsonquery.New(node).ArrayRequired(responseKey)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(arr)
	}
}

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		cursor, err := jsonquery.New(node, "response_metadata").StringOptional("next_cursor")
		if err != nil {
			return "", err
		}

		if cursor == nil || *cursor == "" {
			return "", nil
		}

		return *cursor, nil
	}
}
