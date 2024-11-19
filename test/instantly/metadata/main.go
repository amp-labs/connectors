package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/instantly"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var objectName = "tags" // nolint: gochecknoglobals

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetInstantlyConnector(ctx)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
	})
	if err != nil {
		utils.Fail("error reading from Smartlead", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Smartlead", "error", err)
	}

	slog.Info("Comparing")

	mismatchErr := mockutils.ValidateReadConformsMetadata(objectName, response.Data[0].Raw, metadata)
	if mismatchErr != nil {
		utils.Fail("Failure: Schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		slog.Info("Success: Object metadata schema and endpoint response have the same fields")
	}
}
