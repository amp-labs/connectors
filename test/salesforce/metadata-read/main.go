package main

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var objectName = "Organization" // nolint: gochecknoglobals

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	defer utils.Close(conn)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("*"),
	})
	if err != nil {
		utils.Fail("error reading from Salesforce", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Salesforce", "error", err)
	}

	fmt.Println("Compare object metadata against endpoint response:")

	data := sanitizeReadResponse(response.Data[0].Raw)

	mismatchErr := mockutils.ValidateReadConformsMetadata(strings.ToLower(objectName), data, metadata)
	if mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		fmt.Println("==> success fields match.")
	}
}

func sanitizeReadResponse(response map[string]any) map[string]any {
	// every Salesforce response attached attributes object with type and url of a resource.
	// this attribute field will not appear in metadata response, so we shall remove it.
	crucialFields := make(map[string]any)

	for field, v := range response {
		if field != "attributes" {
			crucialFields[field] = v
		}
	}

	return crucialFields
}
