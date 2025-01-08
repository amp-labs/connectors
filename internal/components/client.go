package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/providers"
)

type ClientComponent struct {
	JSON *common.JSONHTTPClient
	xml  *common.XMLHTTPClient
	ProviderComponent
}

func NewClientComponent(
	provider providers.Provider,
	params common.Parameters,
) (*ClientComponent, error) {
	providerComponent, err := newProviderComponent(provider, params.Module, params.Metadata)
	if err != nil {
		return nil, err
	}

	cc := &ClientComponent{
		JSON: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: interpreter.ErrorHandler{
					JSON: interpreter.NewFaultyResponder(errorFormats(), statusCodeMapping()),
				}.Handle,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
		xml: &common.XMLHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: interpreter.ErrorHandler{
					JSON: interpreter.NewFaultyResponder(errorFormats(), statusCodeMapping()),
				}.Handle,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
		ProviderComponent: *providerComponent,
	}

	cc.JSON.HTTPClient.Base = cc.ProviderInfo().BaseURL
	cc.xml.HTTPClient.Base = cc.ProviderInfo().BaseURL

	return cc, nil
}

func errorFormats() *interpreter.FormatSwitch { return nil }
func statusCodeMapping() map[int]error        { return nil }

func (c *ClientComponent) JSONHTTPClient() *common.JSONHTTPClient { return c.JSON }
func (c *ClientComponent) HTTPClient() *common.HTTPClient         { return c.JSON.HTTPClient }
