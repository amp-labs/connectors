package snowflake

import (
	"context"
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

// Sentinel errors for read operations.
var (
	errObjectNotFound            = errors.New("object not found in connector configuration")
	errObjectNoColumns           = errors.New("object not found or has no columns")
	errStreamNotConfigured       = errors.New("stream.name not configured for object")
	errConsumptionTableNotConfig = errors.New("consumptionTable not configured for object")
	errDynamicTableNotConfig     = errors.New("dynamicTable.name not configured for object")
	errPrimaryKeyRequired        = errors.New("primaryKey is required for consistent pagination")
	errTimestampColumnRequired   = errors.New("timestampColumn is required when Since or Until is specified")
	errObjectsValidationFailed   = errors.New("snowflake objects validation failed")
	errInvalidPathDepth          = errors.New("invalid path depth")
	errInvalidParentKey          = errors.New("invalid parent key")
	errUnknownProperty           = errors.New("unknown property")
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

// NewConnector creates a new Snowflake connector.
//
// TODO:
//   - Error handling.
//   - Figure out permissions needed for write.
//   - Pagination.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.Snowflake, params, constructor)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connector: %w", err)
	}

	if err := connector.setup(context.Background(), params); err != nil {
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

func (c *Connector) Close() error {
	if c.handle == nil {
		return nil
	}

	return c.handle.db.Close()
}

func (c *Connector) setup(ctx context.Context, params common.ConnectorParams) error {
	var err error

	// Parse per-object configurations.
	c.objects, err = newSnowflakeObjects(params.Metadata)
	if err != nil {
		return fmt.Errorf("failed to parse objects: %w", err)
	}

	// Validate that all objects have required configuration.
	if err := c.objects.Validate(); err != nil {
		return err
	}

	// Create connection info.
	c.handle, err = newConnectionInfoFromParams(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create connection info: %w", err)
	}

	return nil
}
