package hubspot

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Connector is a Hubspot connector.
type Connector struct {
	Module  string
	BaseURL string
	Client  *common.JSONHTTPClient
}

// NewConnector returns a new Hubspot connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
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

	params := &hubspotParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadInfo(providers.Hubspot, nil)
	if err != nil {
		return nil, err
	}

	params.client.HTTPClient.Base = providerInfo.BaseURL
	conn = &Connector{
		BaseURL: params.client.HTTPClient.Base,
		Module:  params.module,
		Client:  params.client,
	}

	conn.Client.HTTPClient.ErrorHandler = conn.interpretError

	return conn, nil
}
