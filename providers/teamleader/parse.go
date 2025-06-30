package teamleader

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func records() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		records, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL(req *http.Request) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {

		res, err := jsonquery.New(node).ArrayRequired("data")
		if err != nil {
			return "", err
		}

		if res == nil {
			return "", nil
		}

		if len(res) > 0 {
			reqBody := req.Body
			if reqBody == nil {
				return "", nil
			}

			bodyBytes, err := io.ReadAll(reqBody)
			if err != nil {
				return "", err
			}

			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			var body map[string]any
			if err := json.Unmarshal(bodyBytes, &body); err != nil {
				return "", err
			}

			if offset, ok := body["page"].(map[string]any); ok {
				if number, ok := offset["number"].(int); ok {

					nextPage := number + 1

					return string(nextPage), nil

				}
			}
		}

		return "", nil
	}
}
