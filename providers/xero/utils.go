package xero

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiBasePath, naming.CapitalizeFirstLetter(objName))
	if err != nil {
		return nil, err
	}

	return url, nil
}

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
