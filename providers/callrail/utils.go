package callrail

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

var responseField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint: gochecknoglobals
	"integration_triggers": "integration_criteria",
	"text-messages":        "conversations",
}, func(objectname string) string {
	return objectname
})

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
