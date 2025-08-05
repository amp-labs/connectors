package copper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/copper/internal/metadata"
)

const apiVersion = "v1"

// applicationHeader is required for all REST API calls.
// https://developer.copper.com/introduction/requests.html#headers
var applicationHeader = common.Header{ // nolint:gochecknoglobals
	Key:   "X-PW-Application",
	Value: "developer_api",
}

type Connector struct {
	*components.Connector
	common.RequireAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader

	userEmail string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Copper, params, constructor)
	if err != nil {
		return nil, err
	}

	connector.userEmail = params.Metadata["userEmail"]

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{"userEmail"},
		},
	}

	errorHandler := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, statusCodeMapping),
	}.Handle

	connector.SchemaProvider = schema.NewOpenAPISchemaProvider(connector.ProviderContext.Module(), metadata.Schemas)

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  errorHandler,
		},
	)

	return connector, nil
}

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	objectPath, err := metadata.Schemas.FindURLPath(common.ModuleRoot, objectName)
	if err != nil {
		return nil, err
	}

	return urlbuilder.New(c.ModuleInfo().BaseURL, apiVersion, objectPath)
}

// emailHeader is required for all REST API calls.
// https://developer.copper.com/introduction/requests.html#headers
func (c *Connector) emailHeader() common.Header {
	return common.Header{
		Key:   "X-PW-UserEmail",
		Value: c.userEmail,
	}
}
