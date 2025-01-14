package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type ClientComponent struct {
	JSON *common.JSONHTTPClient
	XML  *common.XMLHTTPClient
	ProviderComponent
}

func NewClientComponent(
	provider providers.Provider,
	params common.Parameters,
) (*ClientComponent, error) {
	providerComponent, err := newProviderComponent(provider, params.Module, params.Workspace, params.Metadata)
	if err != nil {
		return nil, err
	}

	clientComponent := &ClientComponent{
		JSON: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: common.InterpretError,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
		XML: &common.XMLHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: common.InterpretError,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
		ProviderComponent: *providerComponent,
	}

	clientComponent.JSON.HTTPClient.Base = providerComponent.ProviderInfo().BaseURL
	clientComponent.XML.HTTPClient.Base = providerComponent.ProviderInfo().BaseURL

	return clientComponent, nil
}

func (c *ClientComponent) JSONHTTPClient() *common.JSONHTTPClient { return c.JSON }
func (c *ClientComponent) HTTPClient() *common.HTTPClient         { return c.JSON.HTTPClient }
