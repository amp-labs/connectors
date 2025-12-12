package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// Metadata keys for connection configuration.
const (
	metadataKeyDatabase  = "database"
	metadataKeySchema    = "schema"
	metadataKeyRole      = "role"
	metadataKeyWarehouse = "warehouse"
)

// TODO: Manage connection lifecycle.
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
	ctx context.Context,
	db *sql.DB,
	warehouse, database, schema, role string,
) (*connectionInfo, error) {
	connInfo := &connectionInfo{
		db:        db,
		warehouse: warehouse,
		database:  database,
		schema:    schema,
		role:      role,
	}

	if err := connInfo.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate connection info: %w", err)
	}

	if err := connInfo.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return connInfo, nil
}

func newConnectionInfoFromParams(ctx context.Context, params common.ConnectorParams) (*connectionInfo, error) {
	db, ok := params.CustomAuthenticatedClient.(*sql.DB)
	if !ok {
		return nil, errInvalidCustomAuthenticatedClient
	}

	return newConnectionInfo(
		ctx,
		db,
		params.Metadata[metadataKeyWarehouse],
		params.Metadata[metadataKeyDatabase],
		params.Metadata[metadataKeySchema],
		params.Metadata[metadataKeyRole],
	)
}
