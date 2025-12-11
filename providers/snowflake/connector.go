package snowflake

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

var (
	errInvalidCustomAuthenticatedClient = errors.New("invalid custom authenticated client")
	errMissingWarehouse                 = errors.New("missing warehouse")
	errMissingDatabase                  = errors.New("missing database")
	errMissingSchema                    = errors.New("missing schema")
	errMissingRole                      = errors.New("missing role")
)

type Connector struct {
	*components.Connector

	// Required for account
	common.RequireWorkspace

	// Required for query, warehouse, etc.
	common.RequireMetadata

	// Required for preauthenticated sql.DB instance from gosnowflake.
	common.RequireCustomAuthenticatedClient

	// Functionalities that the connector provides.
	components.SchemaProvider
	components.Reader

	// Required to establish / maintain connection.
	handle *connectionInfo

	// Per-object configurations parsed from metadata.
	// Key is objectName (e.g., "contacts__stream").
	objects *Objects
}

// TODO:
// 1. Error handling
// 2. Figure out permissions needed for write
// 3. Pagination
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Snowflake, params, constructor)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connector: %w", err)
	}

	if err := connector.setup(params); err != nil {
		return nil, fmt.Errorf("failed to setup connector: %w", err)
	}

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)
	connector.Reader = reader.NewDelegateReader(connector.Read)

	return connector, nil
}

func (c *Connector) setup(params common.ConnectorParams) error {
	var err error

	// Parse per-object configurations.
	c.objects, err = newSnowflakeObjects(params.Metadata)
	if err != nil {
		return fmt.Errorf("failed to parse objects: %w", err)
	}

	// Create connection info.
	c.handle, err = newConnectionInfoFromParams(params)
	if err != nil {
		return fmt.Errorf("failed to create connection info: %w", err)
	}

	return nil
}
