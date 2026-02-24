package bigquery

import (
	"errors"

	"cloud.google.com/go/bigquery"
	bqstorage "cloud.google.com/go/bigquery/storage/apiv1"
)

var (
	errInvalidAuthClient = errors.New("CustomAuthenticatedClient must be *BigQueryAuth")
	errNilBigQueryClient = errors.New("BigQueryAuth.Client must not be nil")
	errNilStorageClient  = errors.New("BigQueryAuth.StorageClient must not be nil")
)

// BigQueryAuth bundles the pre-authenticated clients the connector needs.
//
// The BigQuery connector requires two clients:
//
//  1. A *bigquery.Client for SQL operations (metadata queries, schema lookups).
//  2. A *bqstorage.BigQueryReadClient for the Storage Read API (parallel streaming).
//
// Both clients are created by the caller (server or test harness) from the same
// service account credentials. The connector never sees raw credentials.
type BigQueryAuth struct {
	// Client is a pre-authenticated BigQuery client for SQL operations
	// (metadata queries, schema lookups, etc.).
	Client *bigquery.Client

	// StorageClient is a pre-authenticated BigQuery Storage Read API client
	// for parallel streaming reads via Arrow format.
	StorageClient *bqstorage.BigQueryReadClient
}

// Validate checks that all required fields are set.
func (a *BigQueryAuth) Validate() error {
	if a.Client == nil {
		return errNilBigQueryClient
	}

	if a.StorageClient == nil {
		return errNilStorageClient
	}

	return nil
}
