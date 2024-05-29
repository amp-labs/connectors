package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/dynamicscrm"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

var (
	objectName     = "accounts"
	objectNameMeta = "account"
)

// we want to compare fields returned by read and schema properties provided by metadata methods
// they must match for all such objects
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	filePath := os.Getenv("MS_CRM_CRED_FILE")
	if filePath == "" {
		filePath = "./ms-crm-creds.json"
	}

	conn := connTest.GetMSDynamics365CRMConnector(ctx, filePath)
	defer utils.Close(conn)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
	})
	if err != nil {
		utils.Fail("error reading from microsoft CRM", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	beforeCall := time.Now()

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectNameMeta,
	})
	if err != nil {
		utils.Fail("error listing metadata for microsoft CRM", "error", err)
	}

	fmt.Printf("ListObjectMetadata took %.2f seconds.\n", time.Since(beforeCall).Seconds())

	fmt.Println("Compare object metadata against endpoint response:")

	data := sanitizeReadResponse(response.Data[0].Raw)

	mismatchErr := mockutils.ValidateReadConformsMetadata(objectNameMeta, data, metadata)
	if mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	} else {
		fmt.Println("==> success fields match.")
	}
}

func sanitizeReadResponse(response map[string]any) map[string]any {
	crucialFields := make(map[string]any)

	for field, v := range response {
		// ignore all fields that are OData annotations
		// they are not part of ObjectMetadata
		if !strings.HasPrefix(field, "@") {
			crucialFields[field] = v
		}
	}

	return crucialFields
}
