package awsic

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/providers"
)

var ErrMissingMetadataIDs = errors.New("missing metadata identifiers")

type Connector struct {
	// Basic connector
	*components.Connector

	// Require authenticated client
	common.RequireAuthenticatedClient

	// supported operations
	components.Reader

	identityStoreID string
	instanceArn string
}

func NewConnector(params common.Parameters) (*Connector, error) {
	// Create base connector with provider info
	conn, err := components.Initialize(providers.AWSIdentityCenter, params, constructor)
	if err != nil {
		return nil, err
	}

	conn.identityStoreID = params.Metadata["IdentityStoreID"]
	conn.instanceArn = params.Metadata["InstanceArn"]

	if len(conn.identityStoreID) == 0 || len(conn.instanceArn) == 0 {
		return nil, ErrMissingMetadataIDs
	}

	return conn, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

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
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
