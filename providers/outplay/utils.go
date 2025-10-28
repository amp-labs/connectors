package outplay

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/amp-labs/connectors/common"
)

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func buildMetadataBody(objectName string) (*bytes.Reader, error) {
	body := map[string]any{
		"pageindex": 1,
	}

	if objectName == "call" {
		now := time.Now()

		// Call object requires fromdate and todate parameters.
		// We set todate as current date and fromdate as 30 days ago to get recent calls.
		thirtyDaysAgo := now.AddDate(0, 0, -30)

		body["fromdate"] = thirtyDaysAgo.Format(timeLayout)
		body["todate"] = now.Format(timeLayout)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(bodyJSON)

	return bodyReader, nil
}
