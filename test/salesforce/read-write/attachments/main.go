package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

// Prerequisites - at least one "Account" object must exist.
// Attachment will be created onto first "Account" instance.
// Then we query attachments that were created in the last 10 seconds.
func main() { //nolint:funlen
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	fmt.Println("Lookup account")

	accountID := readAccount(ctx, conn)

	fmt.Println("Create attachment to account")

	since := time.Now().Add(-10 * time.Second)

	createAttachment(ctx, conn, accountID)
	attachments := readAttachments(ctx, conn, since)

	fmt.Println("Reading attachments")
	utils.DumpJSON(attachments, os.Stdout)
}

func readAccount(ctx context.Context, conn *salesforce.Connector) string {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Account",
		Fields: connectors.Fields(
			"id",
		),
	})
	if err != nil {
		utils.Fail("error reading from Salesforce", "error", err)
	}

	data := res.Data
	if len(data) == 0 {
		utils.Fail("no accounts object in Salesforce", "error", err)
	}

	return data[0].Fields["id"].(string)
}

func createAttachment(ctx context.Context, conn *salesforce.Connector, parentID string) {
	res, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Attachment",
		RecordData: map[string]any{
			"ParentId": parentID,
			"Name":     "attachment coming from the test script",
			"Body":     "awesome text",
		},
	})
	if err != nil {
		utils.Fail("error writing to Salesforce", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create an attachment")
	}
}

func readAttachments(ctx context.Context, conn *salesforce.Connector, since time.Time) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Attachment",
		Fields:     connectors.Fields("id"),
		Since:      since,
	})
	if err != nil {
		utils.Fail("error reading from Salesforce", "error", err)
	}

	return res
}
