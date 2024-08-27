package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var objectName = "issue" // nolint: gochecknoglobals

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAtlassianConnector(ctx)
	defer utils.Close(conn)

	response, err := conn.Read(ctx, common.ReadParams{})
	if err != nil {
		utils.Fail("error reading from Atlassian", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, nil)
	if err != nil {
		utils.Fail("error listing metadata for Atlassian", "error", err)
	}

	slog.Info("Compare object metadata against endpoint response:")

	mismatchErr := mockutils.ValidateReadConformsMetadata(objectName, response.Data[0].Raw, metadata)
	if mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		slog.Info("==> success fields match.")
	}
}
