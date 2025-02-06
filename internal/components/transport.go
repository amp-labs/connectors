package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// TODO: Add support for XML, CSV, etc.
type Transport struct {
	ProviderContext
	json *common.JSONHTTPClient
}

func NewTransport(
	provider providers.Provider,
	params common.Parameters,
) (*Transport, error) {
	providerContext, err := NewProviderContext(provider, params.Module, params.Workspace, params.Metadata)
	if err != nil {
		return nil, err
	}

	return &Transport{
		ProviderContext: *providerContext,
		json: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Base:   providerContext.ProviderInfo().BaseURL,
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: common.InterpretError,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
	}, nil
}

func (c *Transport) JSONHTTPClient() *common.JSONHTTPClient { return c.json }
func (c *Transport) HTTPClient() *common.HTTPClient         { return c.json.HTTPClient }
