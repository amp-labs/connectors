package exensify

import (
	"encoding/json"

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

func buildReadBody(objectName string) (string, error) {

	body := map[string]any{
		"type": "get",
	}

	if objectName == objectNamePolicy {
		body["inputSettings"] = map[string]any{
			"type":      "policyList",
			"adminOnly": true,
		}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(jsonBody), nil
}
