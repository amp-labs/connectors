package bigquery

import (
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/deleter"
	"github.com/amp-labs/connectors/internal/components/reader"
	"github.com/amp-labs/connectors/internal/components/schema"
	"github.com/amp-labs/connectors/internal/components/writer"
	"github.com/amp-labs/connectors/providers"
)

var (
	errInvalidCustomAuthenticatedClient = errors.New("invalid custom authenticated client: expected *bigquery.Client")
	errMissingDataset                   = errors.New("missing dataset in metadata")
	errMissingProject                   = errors.New("missing project in metadata")
	errMissingCredentials               = errors.New("missing credentials in metadata")
	metadataDatasetKey                  = "dataset"
	metadataProjectKey                  = "project"
	metadataCredentialsKey              = "credentials"
)

// Connector provides BigQuery read/write operations using the Storage Read API.
//
// # Authentication
//
// This connector requires a pre-authenticated *bigquery.Client passed via
// CustomAuthenticatedClient. Additionally, the Storage Read API requires
// separate credentials passed via metadata["credentials"] as a JSON string.
//
// # Required Metadata
//
//   - project: GCP project ID
//   - dataset: BigQuery dataset name
//   - credentials: Service account JSON for Storage API authentication
type Connector struct {
	*components.Connector

	common.RequireCustomAuthenticatedClient
	common.RequireMetadata

	components.SchemaProvider
	components.Reader
	components.Writer
	components.Deleter

	// handle is the pre-authenticated BigQuery client for SQL operations.
	handle *bigquery.Client

	// project is the GCP project ID.
	project string

	// dataset is the BigQuery dataset name.
	dataset string

	// credentials is the service account JSON for Storage API authentication.
	// The Storage API requires separate auth from the main BigQuery client.
	credentials []byte
}

// NewConnector creates a new BigQuery connector.
func NewConnector(params common.ConnectorParams) (*Connector, error) {
	connector, err := components.Initialize(providers.BigQuery, params, constructor)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connector: %w", err)
	}

	var ok bool
	connector.handle, ok = params.CustomAuthenticatedClient.(*bigquery.Client)
	if !ok || connector.handle == nil {
		return nil, errInvalidCustomAuthenticatedClient
	}

	connector.project, ok = params.Metadata[metadataProjectKey]
	if !ok || connector.project == "" {
		return nil, errMissingProject
	}

	connector.dataset, ok = params.Metadata[metadataDatasetKey]
	if !ok || connector.dataset == "" {
		return nil, errMissingDataset
	}

	// Credentials are required for the Storage Read API.
	credsStr, ok := params.Metadata[metadataCredentialsKey]
	if !ok || credsStr == "" {
		return nil, errMissingCredentials
	}

	connector.credentials = []byte(credsStr)

	return connector, nil
}

func constructor(base *components.Connector) (*Connector, error) {
	connector := &Connector{Connector: base}

	connector.SchemaProvider = schema.NewDelegateSchemaProvider(connector.listObjectMetadata)
	connector.Reader = reader.NewDelegateReader(connector.Read)
	connector.Writer = writer.NewDelegateWriter(connector.Write)
	connector.Deleter = deleter.NewDelegateDeleter(connector.Delete)

	return connector, nil
}

// Close closes the BigQuery client connection.
func (c *Connector) Close() error {
	if c.handle == nil {
		return nil
	}

	return c.handle.Close()
}
