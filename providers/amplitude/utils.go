package amplitude

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

func (c *Connector) constructURL(objectName string) (*urlbuilder.URL, error) {
	apiVersion := objectAPIVersion.Get(objectName)

	if objectName == objectNameEvents {
		objectName += "/list"
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, "api", apiVersion, objectName)
}
