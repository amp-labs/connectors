package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/providers"
)

// Transport
// TODO: Add support for XML, CSV, etc.
type Transport struct {
	ProviderContext
	RootClient   APIClient
	ModuleClient APIClient
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

	return &Transport{
		ProviderContext: *providerContext,
		RootClient: *NewAPIClient(
			common.ModuleRoot,
			providerContext.providerInfo.BaseURL,
			params.AuthenticatedClient,
			variables,
		),
		ModuleClient: *NewAPIClient(
			providerContext.moduleID,
			providerContext.moduleInfo.BaseURL,
			params.AuthenticatedClient,
			variables,
		),
	}, nil
}

func (t *Transport) SetErrorHandler(handler common.ErrorHandler) {
	t.RootClient.SetErrorHandler(handler)
	t.ModuleClient.SetErrorHandler(handler)
}

func (t *Transport) JSONHTTPClient() *common.JSONHTTPClient {
	return t.RootClient.JSONHTTPClient
}

func (t *Transport) HTTPClient() *common.HTTPClient {
	return t.RootClient.JSONHTTPClient.HTTPClient
}

func createCatalogVariables(params common.Parameters) []catalogreplacer.CatalogVariable {
	metadata := params.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata[catalogreplacer.VariableWorkspace] = params.Workspace

	return paramsbuilder.NewCatalogVariables(metadata)
}
