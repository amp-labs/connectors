package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

// Transport
// TODO: Add support for XML, CSV, etc.
type Transport struct {
	ProviderContext
	json *common.JSONHTTPClient

	ProviderURLBuilder urlbuilder.Template
	ModuleURLBuilder   urlbuilder.Template
}

// NewTransport
// TODO: The JSON client by itself is not providing any functionality right now - this is to only provide
// continuity for the existing codebase. We should refactor the existing JSON/XML/CSV/HTTP clients to
// satisfy a common interface, and then hook them up in here.
func NewTransport(
	provider providers.Provider,
	params common.Parameters,
) (*Transport, error) {
	variables := createCatalogVariables(params)

	providerContext, err := NewProviderContext(provider, params.Module, variables)
	if err != nil {
		return nil, err
	}

	providerBaseURL := providerContext.providerInfo.BaseURL
	moduleBaseURL := providerContext.moduleInfo.BaseURL

	return &Transport{
		ProviderContext: *providerContext,
		json: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Base:   providerBaseURL,
				Client: params.AuthenticatedClient,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: common.InterpretError,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
		ProviderURLBuilder: *urlbuilder.NewTemplate(providerBaseURL),
		ModuleURLBuilder:   *urlbuilder.NewTemplate(moduleBaseURL),
	}, nil
}

func (t *Transport) SetErrorHandler(handler common.ErrorHandler) {
	t.json.HTTPClient.ErrorHandler = handler
}

func (t *Transport) OverrideURL(newURL string) {
	t.ProviderContext.providerInfo.BaseURL = newURL
	t.ProviderContext.moduleInfo.BaseURL = newURL
	t.ProviderURLBuilder.OverrideURL(newURL)
	t.ModuleURLBuilder.OverrideURL(newURL)
}

func (t *Transport) JSONHTTPClient() *common.JSONHTTPClient {
	return t.json
}

func (t *Transport) HTTPClient() *common.HTTPClient {
	return t.json.HTTPClient
}

func createCatalogVariables(params common.Parameters) []catalogreplacer.CatalogVariable {
	metadata := params.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata[catalogreplacer.VariableWorkspace] = params.Workspace

	return paramsbuilder.NewCatalogVariables(metadata)
}
