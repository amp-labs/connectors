package chilipiper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient
}

func NewConnector(opts ...Option) (*Connector, error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	providerInfo, err := providers.ReadInfo(providers.ChiliPiper)
	if err != nil {
		return nil, err
	}

	jsonClient := common.JSONHTTPClient{
		HTTPClient: params.Caller,
	}

	connector := Connector{
		Client: &jsonClient,
	}

	connector.BaseURL = providerInfo.BaseURL

	return &connector, nil
}
