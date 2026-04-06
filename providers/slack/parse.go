package slack

import (
	"errors"
	"fmt"

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
			// Slack usually includes a short error code in the "error" field.
			// Include it in the message if present.
			errorMessage, err := jsonquery.New(node).StringOptional("error")
			if err != nil {
				return nil, err
			}

			if errorMessage != nil {
				return nil, fmt.Errorf("%w %s", errResponseIndicatesFailure, *errorMessage)
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
