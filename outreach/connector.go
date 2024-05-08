package outreach

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	providerOptionRestApiURL = "restAPIURL"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
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

	params := &outreachParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	// Read provider info
	providerInfo, err := providers.ReadInfo(providers.Outreach, nil)
	if err != nil {
		return nil, err
	}

	restApi, ok := providerInfo.GetOption(providerOptionRestApiURL)
	if !ok {
		return nil, fmt.Errorf("restAPIURL not set: %w", providers.ErrProviderOptionNotFound)
	}

	params.client.HTTPClient.Base = providerInfo.BaseURL

	return &Connector{
		Client:  params.client,
		BaseURL: restApi,
	}, nil
}
