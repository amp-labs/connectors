package confluence

import (
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
)

const apiVersion = "v2"

type Adapter struct {
	components.BaseConnector
	components.SchemaProvider
	components.Reader
}

func NewAdapter(base *components.BaseConnector) (*Adapter, error) {
	adapter := &Adapter{
		BaseConnector: *base,
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	adapter.SchemaProvider = schema.NewOpenAPISchemaProvider(adapter.ProviderContext.Module(), Schemas)

	adapter.Reader = reader.NewHTTPReader(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  adapter.buildReadRequest,
			ParseResponse: adapter.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return adapter, nil
}

func (a *Adapter) getRawModuleURL() (*urlbuilder.URL, error) {
	url := strings.ReplaceAll(a.ModuleInfo().BaseURL, "wiki/api", "")

	return urlbuilder.New(url)
}

func (a *Adapter) getReadURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, objectName)
}
