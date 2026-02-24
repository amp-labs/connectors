package bigquery

import (
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	bqstorage "cloud.google.com/go/bigquery/storage/apiv1"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/providers"
)

var (
	errMissingProject         = errors.New("missing project in metadata")
	errMissingDataset         = errors.New("missing dataset in metadata")
	errMissingTimestampColumn = errors.New("missing timestampColumn in metadata: required for incremental reads and backfill windowing")
)

// Connector provides BigQuery read operations using the Storage Read API.
//
// # Authentication
//
// Pass a *BigQueryAuth as CustomAuthenticatedClient. This bundles:
//   - A pre-authenticated *bigquery.Client (for metadata/schema queries)
//   - A pre-authenticated *bqstorage.BigQueryReadClient (for Storage Read API)
//   - A TimestampColumn name (for incremental reads and backfill windowing)
//
// # Required Metadata
//
//   - project: GCP project ID (e.g., "my-gcp-project")
//   - dataset: BigQuery dataset name (e.g., "analytics")
//
// # How reading works
//
// The connector uses the BigQuery Storage Read API for all data reads. This API
// provides parallel streaming via Arrow format, which is significantly faster than
// the SQL query API for large datasets.
//
// For incremental reads (Since/Until set), the connector applies a RowRestriction
// filter on the TimestampColumn. For full backfills (no Since), the connector
// automatically partitions the table into 30-day time windows to ensure each
// Storage API session completes within its 6-hour lifetime. See read.go for details.
type Connector struct {
	*components.Connector

	common.RequireCustomAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader

	// handle is the pre-authenticated BigQuery client for SQL/metadata operations.
	handle *bigquery.Client

	// project is the GCP project ID.
	project string

	// dataset is the BigQuery dataset name.
	dataset string

	// storageClient is the pre-authenticated BigQuery Storage Read API client.
	storageClient *bqstorage.BigQueryReadClient

	// timestampColumn is the column used for time-based filtering and backfill windowing.
	// Must be TIMESTAMP or DATETIME type. Required for all reads.
	timestampColumn string
}

// NewConnector creates a new BigQuery connector.
//
// Example usage:
//
//	auth := &bigquery.BigQueryAuth{
//	    Client:        bqClient,      // pre-authenticated *bigquery.Client
//	    StorageClient: storageClient, // pre-authenticated Storage Read API client
//	}
//
//	conn, err := bigquery.NewConnector(common.ConnectorParams{
//	    CustomAuthenticatedClient: auth,
//	    Metadata: map[string]string{
//	        "project":         "my-gcp-project",
//	        "dataset":         "analytics",
//	        "timestampColumn": "updated_at",
//	    },
//	})
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.BigQuery, params, constructor)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connector: %w", err)
	}

	// Extract and validate the auth wrapper.
	auth, ok := params.CustomAuthenticatedClient.(*BigQueryAuth)
	if !ok || auth == nil {
		return nil, errInvalidAuthClient
	}

	if err := auth.Validate(); err != nil {
		return nil, fmt.Errorf("invalid BigQueryAuth: %w", err)
	}

	connector.handle = auth.Client
	connector.storageClient = auth.StorageClient

	// Extract required metadata.
	connector.project, ok = params.Metadata["project"]
	if !ok || connector.project == "" {
		return nil, errMissingProject
	}

	connector.dataset, ok = params.Metadata["dataset"]
	if !ok || connector.dataset == "" {
		return nil, errMissingDataset
	}

	connector.timestampColumn, ok = params.Metadata["timestampColumn"]
	if !ok || connector.timestampColumn == "" {
		return nil, errMissingTimestampColumn
	}

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)
	connector.Reader = reader.NewDelegateReader(connector.Read)

	return connector, nil
}

// Close closes the BigQuery client connections.
func (c *Connector) Close() error {
	var errs []error

	if c.storageClient != nil {
		if err := c.storageClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.handle != nil {
		if err := c.handle.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
