package main

import (
	"context"
	"log/slog"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var (
	objectName = "Leads" // nolint: gochecknoglobals
)

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPipelinerConnector(ctx)
	defer utils.Close(conn)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
	})
	if err != nil {
		utils.Fail("error reading from Pipeliner", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Pipeliner", "error", err)
	}

	slog.Info("Compare object metadata against endpoint response:")

	data := sanitizeReadResponse(response.Data[0].Raw)

	mismatchErr := mockutils.ValidateReadConformsMetadata(objectName, data, metadata)
	if mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		slog.Info("==> success fields match.")
	}
}

func sanitizeReadResponse(response map[string]any) map[string]any {
	// Pipeliner has some extra fields attached starting with `cf_` prefix.
	// Example Leads:
	//	cf_other_lead_source
	//	cf_lead_source1
	//	cf_lead_source1_id
	crucialFields := make(map[string]any)

	for field, v := range response {
		if !strings.HasPrefix(field, "cf_") {
			crucialFields[field] = v
		}
	}

	return crucialFields
}
