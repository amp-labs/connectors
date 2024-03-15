package proxy

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Connector is a Hubspot connector.
type Connector struct {
	Client *common.JSONHTTPClient

	provider      providers.Provider
	providerInfo  *providers.ProviderInfo
	substitutions *map[string]string
}

func (c *Connector) validate() error {
	return nil
}

type Option func(conn *Connector)

// NewConnector returns a new proxy connector.
func NewConnector(provider providers.Provider, opts ...Option) (*Connector, error) {
	// Initialise the connector
	connector := &Connector{
		provider: provider,
	}

	// Manage panic recovery
	var outErr error

	defer func() {
		if re := recover(); re != nil {
			tmp, ok := re.(error)
			if !ok {
				panic(re)
			}

			outErr = tmp
			connector = nil
		}
	}()

	// Apply options to the connector
	for _, opt := range opts {
		opt(connector)
	}

	// This assumes that the client was set via the options
	if connector.Client == nil {
		return nil, ErrMissingClient
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadInfo(provider, connector.substitutions)
	if err != nil {
		return nil, err
	}

	// Assign substituted provider info to the connector & init client
	connector.providerInfo = providerInfo
	connector.Client.HTTPClient.Base = providerInfo.BaseURL
	connector.Client.HTTPClient.ErrorHandler = connector.interpretError

	// TODO: Does this need to be here?
	// Validate the connector
	outErr = connector.validate()

	return connector, outErr
}

func (c *Connector) JSONClient() *common.JSONHTTPClient {
	return c.Client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.Client.HTTPClient
}
