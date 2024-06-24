package outreach

import (
	"fmt"

	"github.com/amp-labs/connectors/catalog"
	"github.com/amp-labs/connectors/common"
)

const (
	providerOptionRestApiURL = "restAPIURL"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
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

	// Read provider info
	providerInfo, err := catalog.ReadInfo(catalog.Outreach, nil)
	if err != nil {
		return nil, err
	}

	restApi, ok := providerInfo.GetOption(providerOptionRestApiURL)
	if !ok {
		return nil, fmt.Errorf("restAPIURL not set: %w", catalog.ErrProviderOptionNotFound)
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
		BaseURL: restApi,
	}

	conn.Client.HTTPClient.Base = providerInfo.BaseURL

	return conn, nil
}
