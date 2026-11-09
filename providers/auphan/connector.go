package auphan

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

var errInvalidStoreID = errors.New("invalid store ID")

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient
	common.RequireWorkspace

	// Supported operations
	components.SchemaProvider
	components.Reader

	// storeID goes into every request body as "store_id".
	// A franchise HQ token can read several stores by changing store_id per request; for now one
	// connection reads one store.
	storeID int
}

// NewConnector connects to a production store. The workspace is the store ID, which also picks the host.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return newConnector(providers.Auphan, params, params.Workspace)
}

// NewSandboxConnector connects to an Auphan demo install. The workspace is the demo name,
// and the seeded store ID comes from the "storeId" metadata.
func NewSandboxConnector(params common.ConnectorParams) (*Connector, error) {
	return newConnector(providers.AuphanSandbox, params, params.Metadata["storeId"])
}

func newConnector(provider providers.Provider, params common.ConnectorParams, storeID string) (*Connector, error) {
	conn, err := components.Initialize(provider, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.storeID, err = strconv.Atoi(storeID)
	if err != nil {
		return nil, fmt.Errorf("%w: store ID must be a number, got %q", errInvalidStoreID, storeID)
	}

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewObjectSchemaProvider(
		connector.HTTPClient().Client,
		schema.FetchModeParallel,
		operations.SingleObjectMetadataHandlers{
			BuildRequest:  connector.buildSingleObjectMetadataRequest,
			ParseResponse: connector.parseSingleObjectMetadataResponse,
			ErrorHandler:  interpretError,
		},
	)

	registry, err := components.NewEndpointRegistry(supportedOperations())
	if err != nil {
		return nil, err
	}

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		registry,
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  interpretError,
		},
	)

	return connector, nil
}
