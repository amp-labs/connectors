package bigquery

import (
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
)

var (
	errInvalidAuthClient       = errors.New("CustomAuthenticatedClient must be *BigQueryAuth")
	errNilBigQueryClient       = errors.New("BigQueryAuth.Client must not be nil")
	errEmptyCredentials        = errors.New("BigQueryAuth.Credentials must not be empty")
	errInvalidCredentialType  = errors.New("invalid credential type: only service_account keys are accepted")
	errInvalidCredentialsJSON = errors.New("credentials is not valid JSON")
	errMissingTimestampColumn = errors.New("BigQueryAuth.TimestampColumn must not be empty: required for incremental reads and backfill windowing")
)

// BigQueryAuth bundles everything the connector needs for authentication.
//
// The BigQuery connector requires two forms of authentication:
//
//  1. A *bigquery.Client for SQL operations (metadata queries, schema lookups).
//     This client is pre-authenticated by the caller using any method GCP supports
//     (service account JSON, application default credentials, workload identity, etc.).
//
//  2. Raw service account JSON credentials for the Storage Read API. The Storage API
//     uses a separate gRPC transport and cannot share the HTTP-based bigquery.Client.
//     These credentials are validated to ensure they are "service_account" type only,
//     per Google's security guidance on credential configuration injection attacks.
//     See: https://cloud.google.com/docs/authentication/external/credential-security
//
// Both are packaged into this single struct and passed as CustomAuthenticatedClient,
// keeping the connector constructor signature consistent with other Ampersand connectors.
//
// Additionally, a TimestampColumn is required. This column serves two purposes:
//   - Incremental reads: filters rows with Since/Until via RowRestriction.
//   - Backfill windowing: partitions large tables into time-bounded chunks so that
//     each Storage API session completes well within its 6-hour lifetime.
type BigQueryAuth struct {
	// Client is a pre-authenticated BigQuery client for SQL operations
	// (metadata queries, schema lookups, etc.).
	Client *bigquery.Client

	// Credentials is the raw service account JSON key file content.
	// Used to create Storage Read API gRPC clients. Must have type "service_account".
	Credentials []byte

	// TimestampColumn is the name of the column used for time-based filtering
	// and backfill windowing (e.g., "updated_at", "created_at", "_PARTITIONTIME").
	//
	// This column must exist on every table that will be read. It is used to:
	//   - Apply Since/Until filters as RowRestriction WHERE clauses
	//   - Partition backfills into time windows to avoid session expiry
	//
	// The column must be of type TIMESTAMP or DATETIME in BigQuery.
	TimestampColumn string
}

// Validate checks that all required fields are set and credentials are safe to use.
func (a *BigQueryAuth) Validate() error {
	if a.Client == nil {
		return errNilBigQueryClient
	}

	if len(a.Credentials) == 0 {
		return errEmptyCredentials
	}

	if err := validateServiceAccountCredentials(a.Credentials); err != nil {
		return err
	}

	if a.TimestampColumn == "" {
		return errMissingTimestampColumn
	}

	return nil
}

// validateServiceAccountCredentials checks that the provided credential JSON
// is a service account key (type: "service_account").
//
// Why this matters: Google's credential configuration format allows "external_account"
// types that can point to arbitrary token-exchange endpoints. If an attacker controls
// the credentials input, they could supply an external_account config pointing to their
// server, causing the connector to send authentication requests to a malicious endpoint.
// By rejecting everything except "service_account", we eliminate this attack vector.
//
// See: https://cloud.google.com/docs/authentication/external/credential-security
func validateServiceAccountCredentials(creds []byte) error {
	var parsed struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(creds, &parsed); err != nil {
		return fmt.Errorf("%w: %w", errInvalidCredentialsJSON, err)
	}

	if parsed.Type != "service_account" {
		return fmt.Errorf("%w: got %q", errInvalidCredentialType, parsed.Type)
	}

	return nil
}
