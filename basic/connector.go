package basic

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Connector is a Hubspot connector.
type Connector struct {
	ProviderInfo *providers.ProviderInfo
	Client       *common.JSONHTTPClient
	provider     providers.Provider
}

// NewConnector returns a new Hubspot connector.
func NewConnector(
	provider providers.Provider,
	opts ...Option,
) (conn *Connector, outErr error) {
	defer func() {
		if re := recover(); re != nil {
			tmp, ok := re.(error)
			if !ok {
				panic(re)
			}

			outErr = tmp
			conn = nil
		}
	}()

	// Set up basic params
	params := &basicParams{}
	params.provider = provider

	// Apply options & verify
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	// Create connector
	conn = &Connector{
		provider: params.provider,
		Client:   params.client,
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadConfig(conn.provider, &params.substitutions)
	if err != nil {
		return nil, err
	}

	// Set provider info & http client options
	conn.ProviderInfo = providerInfo
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError

	// Set base URL
	conn.Client.HTTPClient.Base = conn.ProviderInfo.BaseURL

	return conn, nil
}
