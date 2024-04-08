package outreach

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	// we need to change the BaseURL
	// for this to be v2
	apiVersion = "api/v2"
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

	var params = &outreachParams{}
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

	params.client.HTTPClient.Base = providerInfo.BaseURL

	return &Connector{
		Client:  params.client,
		BaseURL: strings.Join([]string{params.client.HTTPClient.Base, apiVersion}, "/"),
	}, nil
}
