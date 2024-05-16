package docusign

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	ProviderInfo *providers.ProviderInfo
	Client       *common.JSONHTTPClient
}

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

	params := &docusignParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: params.client,
	}

	// Read provider info
	conn.ProviderInfo, err = providers.ReadInfo(providers.Docusign, nil)
	if err != nil {
		return nil, err
	}

	// Set the base URL
	conn.Client.HTTPClient.Base = conn.ProviderInfo.BaseURL

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Docusign
}
