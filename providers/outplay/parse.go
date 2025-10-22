package outplay

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

func extractMetadataRecords(res map[string]any, objectName string) ([]any, error) {
	if objectName == "callanalysis" {
		data, ok := res["data"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the response field 'data' to a map: %w", common.ErrMissingExpectedValues)
		}

		records, ok := data["data"].([]any)
		if !ok {
			return nil, fmt.Errorf("couldn't convert the nested response field 'data' to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
		}

		return records, nil
	}

	records, ok := res["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("couldn't convert the response field 'data' to an array: %w", common.ErrMissingExpectedValues) // nolint:lll
	}

	return records, nil
}
