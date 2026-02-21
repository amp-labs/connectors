package bigquery

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/amp-labs/connectors/common"
	connectorbq "github.com/amp-labs/connectors/providers/bigquery"
	"github.com/amp-labs/connectors/test/utils"
	"google.golang.org/api/option"
)

// Environment variables for BigQuery configuration.
const (
	EnvServiceAccountPath = "BIGQUERY_SERVICE_ACCOUNT_PATH"
	EnvProject            = "BIGQUERY_PROJECT"
	EnvDataset            = "BIGQUERY_DATASET"
	EnvLocation           = "BIGQUERY_LOCATION" // optional
)

// GetBigQueryConnector creates a BigQuery connector for testing.
// Required environment variables:
//   - BIGQUERY_SERVICE_ACCOUNT_PATH: Path to service account JSON file
//   - BIGQUERY_PROJECT: GCP project ID
//   - BIGQUERY_DATASET: BigQuery dataset name
//
// Optional environment variables:
//   - BIGQUERY_LOCATION: Dataset location (e.g., "US", "EU")
func GetBigQueryConnector(ctx context.Context) *connectorbq.Connector {
	project := os.Getenv(EnvProject)
	if project == "" {
		project = "ampersand-dev" // Default for local testing.
	}

	dataset := os.Getenv(EnvDataset)
	if dataset == "" {
		dataset = "patents" // Default for local testing.
	}

	credsFile := os.Getenv(EnvServiceAccountPath)
	if credsFile == "" {
		credsFile = "providers/ampersand-dev-d8ceed67df75.json" // Default for local testing.
	}

	serviceAccountJSON, err := os.ReadFile(credsFile)
	if err != nil {
		utils.Fail("error reading service account file", "path", credsFile, "error", err)
	}

	// Create BigQuery client.
	client, err := bigquery.NewClient(ctx, project, option.WithCredentialsJSON(serviceAccountJSON))
	if err != nil {
		utils.Fail("error creating BigQuery client", "error", err)
	}

	// Create connector with required metadata.
	// The connector needs project, dataset, and credentials for Storage API.
	conn, err := connectorbq.NewConnector(common.ConnectorParams{
		CustomAuthenticatedClient: client,
		Metadata: map[string]string{
			"project":     project,
			"dataset":     dataset,
			"credentials": string(serviceAccountJSON),
		},
	})
	if err != nil {
		utils.Fail("error creating BigQuery connector", "error", err)
	}

	return conn
}

// GetBigQueryClient returns the underlying BigQuery client for direct operations.
func GetBigQueryClient(ctx context.Context) *bigquery.Client {
	serviceAccountPath := os.Getenv(EnvServiceAccountPath)
	if serviceAccountPath == "" {
		utils.Fail("missing environment variable", "variable", EnvServiceAccountPath)
	}

	project := os.Getenv(EnvProject)
	if project == "" {
		utils.Fail("missing environment variable", "variable", EnvProject)
	}

	serviceAccountJSON, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		utils.Fail("error reading service account file", "path", serviceAccountPath, "error", err)
	}

	client, err := bigquery.NewClient(ctx, project, option.WithCredentialsJSON(serviceAccountJSON))
	if err != nil {
		utils.Fail("error creating BigQuery client", "error", err)
	}

	return client
}

// PrintUsage prints the required environment variables.
func PrintUsage() {
	fmt.Println("BigQuery Test Configuration")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Required environment variables:")
	fmt.Printf("  %s: Path to service account JSON file\n", EnvServiceAccountPath)
	fmt.Printf("  %s: GCP project ID\n", EnvProject)
	fmt.Printf("  %s: BigQuery dataset name\n", EnvDataset)
	fmt.Println()
	fmt.Println("Optional environment variables:")
	fmt.Printf("  %s: Dataset location (e.g., 'US', 'EU')\n", EnvLocation)
	fmt.Println()
	fmt.Println("Example:")
	fmt.Printf("  export %s=/path/to/service-account.json\n", EnvServiceAccountPath)
	fmt.Printf("  export %s=my-gcp-project\n", EnvProject)
	fmt.Printf("  export %s=my_dataset\n", EnvDataset)
	fmt.Printf("  export %s=US\n", EnvLocation)
}
