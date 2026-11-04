package databricks

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

const (
	metadataKeyWarehouseID     = "warehouseId"
	metadataKeyCatalog         = "catalog"
	metadataKeySchema          = "schema"
	metadataKeyPrimaryKey      = "primaryKey"
	metadataKeyTimestampColumn = "timestampColumn"
)

var (
	errStatementNotSucceeded = errors.New("databricks statement did not succeed")
	errObjectNotFound        = errors.New("table not found or has no columns")
	errInvalidNextPage       = errors.New("invalid NextPage, expected a numeric offset")
	errTimestampRequired     = errors.New("timestampColumn is required to filter by Since or Until")
)

// Connector reads Delta tables through the Databricks SQL Statement Execution API.
//
// An object is a table inside the configured catalog and schema. Rows are read with
// a SELECT ordered by the timestamp and primary key columns, paged with LIMIT and
// OFFSET. Since and Until become bounds on the timestamp column.
//
// The access token only needs SELECT on the tables and CAN USE on the SQL warehouse.
//
// The workspace host (without scheme) is the connector's Workspace.
//
// Required metadata:
//   - warehouseId: SQL warehouse that runs the statements.
//   - catalog, schema: where the tables live.
//   - primaryKey: column that uniquely identifies a row.
//   - timestampColumn: TIMESTAMP column that moves forward on every change.
type Connector struct {
	*components.Connector

	common.RequireAuthenticatedClient
	common.RequireWorkspace
	common.RequireMetadata

	components.SchemaProvider
	components.Reader

	warehouseID     string
	catalog         string
	schema          string
	primaryKey      string
	timestampColumn string
}

func NewConnector(params common.ConnectorParams) (*Connector, error) {
	return components.Init(providers.Databricks, params, constructor)
}

func constructor(params common.ConnectorParams, base *components.Connector) (*Connector, error) {
	connector := &Connector{
		Connector: base,
		RequireMetadata: common.RequireMetadata{
			ExpectedMetadataKeys: []string{
				metadataKeyWarehouseID,
				metadataKeyCatalog,
				metadataKeySchema,
				metadataKeyPrimaryKey,
				metadataKeyTimestampColumn,
			},
		},
		warehouseID:     params.Metadata[metadataKeyWarehouseID],
		catalog:         params.Metadata[metadataKeyCatalog],
		schema:          params.Metadata[metadataKeySchema],
		primaryKey:      params.Metadata[metadataKeyPrimaryKey],
		timestampColumn: params.Metadata[metadataKeyTimestampColumn],
	}

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)

	connector.Reader = reader.NewHTTPReader(
		connector.HTTPClient().Client,
		components.NewEmptyEndpointRegistry(),
		connector.ProviderContext.Module(),
		operations.ReadHandlers{
			BuildRequest:  connector.buildReadRequest,
			ParseResponse: connector.parseReadResponse,
			ErrorHandler:  common.InterpretError,
		},
	)

	return connector, nil
}
