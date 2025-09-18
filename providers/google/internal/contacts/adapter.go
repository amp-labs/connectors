package contacts

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v1"

type Adapter struct {
	*components.Connector
	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Google, params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
		HTML: interpreter.DirectFaultyResponder{Callback: adapter.interpretHTMLError},
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

	adapter.Writer = writer.NewHTTPWriter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  adapter.buildWriteRequest,
			ParseResponse: adapter.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	adapter.Deleter = deleter.NewHTTPDeleter(
		adapter.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		adapter.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  adapter.buildDeleteRequest,
			ParseResponse: adapter.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return adapter, nil
}

func (a *Adapter) getReadURL(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := Schemas.FindURLPath(providers.ModuleGoogleContacts, objectName)
	if err != nil {
		return nil, err
	}

	return a.getURL(objectPath)
}

func (a *Adapter) getURL(path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, path)
}

// resourceIdentifierFormat breaks resourceName into parts, where format is: "name/identifier".
func resourceIdentifierFormat(resourceName string) (recordID string, ok bool) {
	parts := strings.Split(resourceName, "/")
	if len(parts) != 2 { // nolint:mnd
		return "", false
	}

	return parts[1], true
}
