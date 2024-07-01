package docusign

import (
	"github.com/amp-labs/connectors/catalog"
	"github.com/amp-labs/connectors/common"
)

type Connector struct {
	ProviderInfo *catalog.ProviderInfo
	Client       *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := parameters{}.FromOptions(opts...)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	// Read provider info
	conn.ProviderInfo, err = catalog.ReadInfo(catalog.Docusign, nil)
	if err != nil {
		return nil, err
	}

	// Set the base URL
	conn.Client.HTTPClient.Base = conn.ProviderInfo.BaseURL

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() catalog.Provider {
	return catalog.Docusign
}
