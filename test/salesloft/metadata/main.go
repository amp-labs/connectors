package main

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
)

var objectNamePlural = "groups"

// we want to compare fields returned by read and schema properties provided by metadata methods
// they must match for all such objects
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesloftConnector(ctx)
	defer utils.Close(conn)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectNamePlural,
	})
	if err != nil {
		utils.Fail("error reading from Salesloft", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectNamePlural,
	})
	if err != nil {
		utils.Fail("error listing metadata for Salesloft", "error", err)
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

	for name := range metadata.Result[objectNamePlural].FieldsMap {
		if _, found := fields[name]; found {
			fields[name] = true
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
