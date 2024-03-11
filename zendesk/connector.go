package zendesk

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	providerOptionRestApiURL = "restApiUrl"
	providerOptionDomain     = "domain"
)

// Connector is a Zendesk connector.
type Connector struct {
	Domain  string
	BaseURL string
	Client  *common.JSONHTTPClient
}

// NewConnector returns a new Zendesk connector.
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

	params := &zendeskParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error

	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadInfo(providers.Zendesk, &map[string]string{
		"workspace": params.workspace,
	})
	if err != nil {
		return nil, err
	}

	restApi, ok := providerInfo.GetOption(providerOptionRestApiURL)
	if !ok {
		return nil, fmt.Errorf("restApiUrl not set: %w", providers.ErrProviderOptionNotFound)
	}

	domain, ok := providerInfo.GetOption(providerOptionDomain)
	if !ok {
		return nil, fmt.Errorf("domain not set: %w", providers.ErrProviderOptionNotFound)
	}

	conn = &Connector{
		BaseURL: restApi,
		Domain:  domain,
		Client:  params.client,
	}

	conn.Client.HTTPClient.Base = providerInfo.BaseURL
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError

	return conn, nil
}
