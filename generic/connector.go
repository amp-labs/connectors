package generic

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	ProviderInfo *providers.ProviderInfo
	Client       *common.JSONHTTPClient
	provider     providers.Provider
}

func NewConnector(
	provider providers.Provider,
	opts ...Option,
) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{provider: provider}, opts)
	if err != nil {
		return nil, err
	}

	// Create connector
	conn = &Connector{
		provider: params.provider,
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadInfo(conn.provider, params.GetCatalogVars()...)
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
