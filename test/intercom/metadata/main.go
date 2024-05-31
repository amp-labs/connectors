package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	msTest "github.com/amp-labs/connectors/test/intercom"
	"github.com/amp-labs/connectors/test/utils"
)

var (
	objectName = "admins"
)

// we want to compare fields returned by read and schema properties provided by metadata methods
// they must match for all such objects
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("INTERCOM_CRED_FILE")
	if filePath == "" {
		filePath = "./intercom-creds.json"
	}

	conn := msTest.GetIntercomConnector(ctx, filePath)
	defer utils.Close(conn)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
	})
	if err != nil {
		utils.Fail("error reading from Intercom", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Intercom", "error", err)
	}

	fmt.Println("Compare object metadata against endpoint response:")

	mismatchErr := compareFieldsMatch(metadata, response.Data[0].Raw)
	if mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		fmt.Println("==> success fields match.")
	}
}

func compareFieldsMatch(metadata *common.ListObjectMetadataResult, response map[string]any) error {
	fields := make(map[string]bool)

	for field := range response {
		fields[field] = false
	}

	mismatch := make([]error, 0)

	for _, displayName := range metadata.Result[objectName].FieldsMap {
		if _, found := fields[displayName]; found {
			fields[displayName] = true
		}
	}

	// every field from Read must be known to ListObjectMetadata
	for name, found := range fields {
		if !found {
			mismatch = append(mismatch, fmt.Errorf("metadata schema is missing field %v", name))
		}
	}

	return errors.Join(mismatch...)
}
