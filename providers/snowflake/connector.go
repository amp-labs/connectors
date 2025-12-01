package snowflake

import (
	"database/sql"
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
	db        *sql.DB
	warehouse string
	database  string
	schema    string
	role      string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	// TODO:
	// 1. Error handling
	// 2. Manage connection lifecycle
	// 3. Figure out permissions needed for write
	// 4. Pagination
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

// Close closes the database connection.
func (c *Connector) Close() error {
	if c.db != nil {
		return c.db.Close()
	}

	return nil
}

func (c *Connector) setup(params common.ConnectorParams) error {
	var ok bool

	c.db, ok = params.CustomAuthenticatedClient.(*sql.DB)
	if !ok {
		return errInvalidCustomAuthenticatedClient
	}

	c.warehouse, ok = params.Metadata["warehouse"]
	if !ok {
		return errMissingWarehouse
	}

	c.database, ok = params.Metadata["database"]
	if !ok {
		return errMissingDatabase
	}

	c.schema, ok = params.Metadata["schema"]
	if !ok {
		return errMissingSchema
	}

	c.role, ok = params.Metadata["role"]
	if !ok {
		return errMissingRole
	}

	return nil
}
