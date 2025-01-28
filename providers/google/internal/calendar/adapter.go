package calendar

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

const apiVersion = "v3"

type Adapter struct {
	*components.Connector
	components.SchemaProvider
	components.Reader
}

func NewAdapter(params common.ConnectorParams) (*Adapter, error) {
	return components.Initialize(providers.Google, params, constructor)
}

func constructor(base *components.Connector) (*Adapter, error) {
	adapter := &Adapter{
		Connector: base,
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
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

	return adapter, nil
}

func (a *Adapter) getURL(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := Schemas.FindURLPath(providers.ModuleGoogleCalendar, objectName)
	if err != nil {
		return nil, err
	}

	// Primary keyword can be put in place of calendar ID to refer to the user's primary calendar.
	// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
	objectPath = strings.ReplaceAll(objectPath, "{calendarId}", "primary")

	return urlbuilder.New(a.ModuleInfo().BaseURL, apiVersion, objectPath)
}
