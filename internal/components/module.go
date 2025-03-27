package components

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

type APIClient struct {
	*common.JSONHTTPClient

	ID          common.ModuleID
	catalogVars []catalogreplacer.CatalogVariable
}

func NewAPIClient(
	moduleID common.ModuleID, url string, client common.AuthenticatedHTTPClient,
	catalogVars []catalogreplacer.CatalogVariable,
) *APIClient {
	return &APIClient{
		JSONHTTPClient: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Base:   url,
				Client: client,

				// ErrorHandler is set to a default, but can be overridden using options.
				ErrorHandler: common.InterpretError,

				// No ResponseHandler is set, but can be overridden using options.
			},
			ErrorPostProcessor: common.ErrorPostProcessor{},
		},
		ID:          moduleID,
		catalogVars: catalogVars,
	}
}

func (c APIClient) URL(path ...string) (*urlbuilder.URL, error) {
	return c.TemplateURL(nil, path...)
}

func (c APIClient) SetURL(newURL string) {
	c.JSONHTTPClient.HTTPClient.Base = newURL
}

func (c APIClient) TemplateURL(templateVars map[string]string, path ...string) (*urlbuilder.URL, error) {
	catalogVars := paramsbuilder.NewCatalogVariables(templateVars)
	catalogVars = append(catalogVars, c.catalogVars...)
	registry := catalogreplacer.NewCatalogSubstitutionRegistry(catalogVars)
	template := &struct{ Data string }{
		Data: c.JSONHTTPClient.HTTPClient.Base,
	}

	if err := registry.Apply(template); err != nil {
		return nil, err
	}

	return urlbuilder.New(template.Data, path...)
}

func (c APIClient) SetErrorHandler(handler common.ErrorHandler) {
	c.JSONHTTPClient.HTTPClient.ErrorHandler = handler
}
