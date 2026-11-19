package auphan

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	_ "time/tzdata" // store timezones resolve on hosts without zoneinfo

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
	common.RequireAuthenticatedClient
	common.RequireWorkspace
	common.RequireMetadata
	components.SchemaProvider
	components.Reader

	// storeID goes into every request body as "store_id".
	// A franchise HQ token can read several stores by changing store_id per request; for now one
	// connection reads one store.
	storeId int

	// location is the store timezone from the "timezone" metadata, UTC when not set.
	// Report times are converted to it. Auphan doesn't return a store's timezone.
	location *time.Location
}

// NewConnector connects to a production store. The workspace is the store ID, which also picks the host.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return newConnector(providers.Auphan, params)
}

// NewSandboxConnector connects to an Auphan demo install. The workspace is the store ID seeded in
// the demo, and the "demoName" metadata picks the demo install.
func NewSandboxConnector(params common.ConnectorParams) (*Connector, error) {
	return newConnector(providers.AuphanSandbox, params)
}

func newConnector(provider providers.Provider, params common.ConnectorParams) (*Connector, error) {
	conn, err := components.Init(provider, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.storeId, err = strconv.Atoi(params.Workspace)
	if err != nil {
		return nil, fmt.Errorf("%w: store ID must be a number, got %q", errInvalidStoreID, params.Workspace)
	}

	conn.location, err = loadLocation(params.Metadata["timezone"])
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func constructor(_ common.ConnectorParams, base *components.Connector) (*Connector, error) {
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

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  interpretError,
		},
	)

	return connector, nil
}
