package freshdesk

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const restAPIPrefix = "api/v2"

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
	return components.Initialize(providers.Freshdesk, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	return &Connector{Connector: base}, nil
}

func (conn *Connector) getAPIURL(objectName string) (*urlbuilder.URL, error) {
	path := objectResourcePath.Get(objectName)

	return urlbuilder.New(conn.ProviderInfo().BaseURL, restAPIPrefix, path)
}
