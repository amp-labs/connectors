package databricks

import (
	"errors"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

const (
	metadataKeyHTTPPath        = "httpPath"
	metadataKeyCatalog         = "catalog"
	metadataKeySchema          = "schema"
	metadataKeyPrimaryKey      = "primaryKey"
	metadataKeyTimestampColumn = "timestampColumn"
)

var (
	errStatementNotSucceeded = errors.New("databricks statement did not succeed")
	errObjectNotFound        = errors.New("table not found, has no columns, or is not visible to the service principal")
	errInvalidNextPage       = errors.New("invalid NextPage, expected a numeric offset")
	errTimestampRequired     = errors.New("timestampColumn is required to filter by Since or Until")
)

// Connector reads Delta tables through the Databricks SQL Statement Execution API.
//
// An object is a table inside the configured catalog and schema. Rows are read with
// a SELECT ordered by the timestamp and primary key columns, paged with LIMIT and
// OFFSET. Since and Until become bounds on the timestamp column.
//
// The service principal behind the OAuth client credentials only needs SELECT on the tables
// (plus USE CATALOG and USE SCHEMA) and CAN USE on the SQL warehouse.
//
// The workspace host (without scheme) is the connector's Workspace.
//
// Required metadata:
//   - httpPath: HTTP path of the SQL warehouse that runs the statements, as shown in its
//     connection details (/sql/1.0/warehouses/<id>). The bare warehouse ID also works.
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
				metadataKeyHTTPPath,
				metadataKeyCatalog,
				metadataKeySchema,
				metadataKeyPrimaryKey,
				metadataKeyTimestampColumn,
			},
		},
		warehouseID:     WarehouseIDFromHTTPPath(params.Metadata[metadataKeyHTTPPath]),
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

// WarehouseIDFromHTTPPath returns the warehouse ID from the HTTP path shown in the warehouse's
// connection details, e.g. /sql/1.0/warehouses/7c806f46d14e255f. A bare ID is returned as is.
func WarehouseIDFromHTTPPath(httpPath string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(httpPath), "/;")

	return trimmed[strings.LastIndex(trimmed, "/")+1:]
}
