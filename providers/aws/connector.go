package aws

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireModule
	common.RequireMetadata

	// supported operations
	components.Reader
	components.Writer
	components.Deleter

	region          string
	identityStoreId string
	instanceARN     string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Initialize(providers.AWS, params,
		func(connector *components.Connector) (*Connector, error) {
			var expectedMetadataKeys []string
			if params.Module == providers.ModuleAWSIdentityCenter {
				expectedMetadataKeys = []string{"region", "identityStoreId", "instanceARN"}
			}

			return constructor(connector, expectedMetadataKeys)
		},
	)
	if err != nil {
		return nil, err
	}

	conn.region = params.Metadata["region"]
	conn.identityStoreId = params.Metadata["identityStoreId"]
	conn.instanceARN = params.Metadata["instanceARN"]

	return conn, nil
}

func constructor(base *components.Connector, expectedMetadataKeys []string) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireModule: common.RequireModule{
			ExpectedModules: []common.ModuleID{
				providers.ModuleAWSIdentityCenter,
			},
		},
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: expectedMetadataKeys,
		},
	}

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	errorHandler := interpreter.ErrorHandler{
		Custom: map[interpreter.Mime]interpreter.FaultyResponseHandler{
			core.Mime: interpreter.NewFaultyResponder(errorFormats, nil),
		},
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

// nolint:unused
func (c *Connector) getModuleURL() (string, error) {
	modules := c.ProviderInfo().Modules
	if modules == nil {
		return "", common.ErrInvalidModuleDeclaration
	}

	baseURL := (*modules)[c.Module()].BaseURL
	baseURL = strings.Replace(baseURL, "{{.region}}", c.region, 1)

	return baseURL, nil
}
