package snowflake

import (
	"database/sql"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var (
	metadataKeyDatabase  = "database"
	metadataKeySchema    = "schema"
	metadataKeyRole      = "role"
	metadataKeyWarehouse = "warehouse"
)

// TODO: Manage connection lifecycle
type connectionInfo struct {
	db        *sql.DB
	warehouse string
	database  string
	schema    string
	role      string
}

func (a *connectionInfo) validate() error {
	if a.db == nil {
		return errMissingDatabase
	}

	if a.warehouse == "" {
		return errMissingWarehouse
	}

	if a.database == "" {
		return errMissingDatabase
	}

	if a.schema == "" {
		return errMissingSchema
	}

	if a.role == "" {
		return errMissingRole
	}

	return nil
}

func newConnectionInfo(
	db *sql.DB,
	warehouse string,
	database string,
	schema string,
	role string) (*connectionInfo, error) {
	c := &connectionInfo{
		db:        db,
		warehouse: warehouse,
		database:  database,
		schema:    schema,
		role:      role,
	}

	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate connection info: %w", err)
	}

	if err := c.db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return c, nil
}

func newConnectionInfoFromParams(params common.ConnectorParams) (*connectionInfo, error) {
	return newConnectionInfo(
		params.CustomAuthenticatedClient.(*sql.DB),
		params.Metadata[metadataKeyWarehouse],
		params.Metadata[metadataKeyDatabase],
		params.Metadata[metadataKeySchema],
		params.Metadata[metadataKeyRole],
	)
}

func (c *connectionInfo) close() error {
	if c.db != nil {
		return c.db.Close()
	}

	return nil
}
