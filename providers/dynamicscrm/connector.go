package dynamicscrm

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v9.2"

type Connector struct {
	// Basic connector
	*components.Connector

	metadataDiscoveryRepository metadataDiscoveryRepository
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
	return components.Initialize(providers.DynamicsCRM, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	base.SetErrorHandler(interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle)

	conn := &Connector{Connector: base}

	conn.metadataDiscoveryRepository = metadataDiscoveryRepository{
		client:   conn.JSONHTTPClient(),
		buildURL: conn.getURL,
	}

	return conn, nil
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return constructURL(c.ProviderInfo().BaseURL, apiVersion, arg)
}
