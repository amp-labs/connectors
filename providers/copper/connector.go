package copper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
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

	components.Reader
	components.Writer
	components.Deleter

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

	connector.Writer = writer.NewHTTPWriter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.WriteHandlers{
			BuildRequest:  connector.buildWriteRequest,
			ParseResponse: connector.parseWriteResponse,
			ErrorHandler:  errorHandler,
		},
	)

	// Set the deleter for the connector
	connector.Deleter = deleter.NewHTTPDeleter(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.DeleteHandlers{
			BuildRequest:  connector.buildDeleteRequest,
			ParseResponse: connector.parseDeleteResponse,
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

func (c *Connector) getWriteURL(objectName string, id string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, id)
}

func (c *Connector) getCustomFieldsURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "custom_field_definitions")
}

// emailHeader is required for all REST API calls.
// https://developer.copper.com/introduction/requests.html#headers
func (c *Connector) emailHeader() common.Header {
	return common.Header{
		Key:   "X-PW-UserEmail",
		Value: c.userEmail,
	}
}
