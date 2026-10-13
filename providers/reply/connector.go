package reply

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

// apiVersion is the URL prefix shared by every Reply V3 endpoint.
// https://docs.reply.io/api-reference
const apiVersion = "v3"

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient

	components.Reader
	components.Writer
	components.Deleter
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Initialize(providers.Reply, params, constructor)
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
	}

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return connector, nil
}
