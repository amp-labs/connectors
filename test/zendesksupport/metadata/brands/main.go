package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	connTest "github.com/amp-labs/connectors/test/zendesksupport"
)

var objectName = "brands" // nolint: gochecknoglobals

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZendeskSupportConnector(ctx)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("name", "time_zone", "role"),
	})
	if err != nil {
		utils.Fail("error reading from ZendeskSupport", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for ZendeskSupport", "error", err)
	}

	slog.Info("Compare object metadata against endpoint response:")

	mismatchErr := mockutils.ValidateReadConformsMetadata(objectName, response.Data[0].Raw, metadata)
	if mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		slog.Info("==> success fields match.")
	}
}
