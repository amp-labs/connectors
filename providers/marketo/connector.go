package marketo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	// Basic connector
	*components.Connector
}

// NewConnector is an old constructor, use NewConnectorV2.
// Deprecated.
func NewConnector(opts ...Option) (*Connector, error) {
	params, err := newParams(opts)
	if err != nil {
		return nil, err
	}

	return NewConnectorV2(*params)
}

func NewConnectorV2(params common.Parameters) (*Connector, error) {
	return components.Initialize(providers.Marketo, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	base.HTTPClient().ResponseHandler = responseHandler

	return &Connector{Connector: base}, nil
}

func (c *Connector) getAPIURL(objName string) (*urlbuilder.URL, error) {
	objName = common.AddSuffixIfNotExists(objName, ".json")

	return c.ModuleClient.URL(objName)
}
