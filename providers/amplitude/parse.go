package amplitude

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// There is no pagination implemented in Amplitude API as of now.
		// So, returning empty string to indicate no next page.
		// Reference: https://amplitude.com/docs/apis/analytics/chart-annotations#get-all-chart-annotations
		return "", nil
	}
}
