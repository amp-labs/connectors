package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/supersend"
)

func main() {
	ctx := context.Background()

	conn := supersend.GetSuperSendConnector(ctx)

	// Test all 12 SuperSend objects

	// 1. teams (standard array response)
	readObject(ctx, conn, "teams", connectors.Fields("id", "name", "domain", "isDefault"))

	// 2. senders (standard array response)
	readObject(ctx, conn, "senders", connectors.Fields("id", "email", "warm", "max_per_day"))

	// 3. sender-profiles
	readObject(ctx, conn, "sender-profiles", connectors.Fields("id", "name", "type", "status"))

	// 4. labels
	readObject(ctx, conn, "labels", connectors.Fields("id", "name", "color", "deleted"))

	// 5. contact/all
	readObject(ctx, conn, "contact/all", connectors.Fields("id", "email", "first_name", "last_name", "status"))

	// 6. campaigns/overview
	readObject(ctx, conn, "campaigns/overview", connectors.Fields("id", "name", "status", "contactedCount"))

	// 7. org (single object response - empty responseKey)
	readObject(ctx, conn, "org", connectors.Fields("id", "name", "current_plan", "domain"))

	// 8. managed-domains
	readObject(ctx, conn, "managed-domains", connectors.Fields("id", "name", "status", "computed_status"))

	// 9. managed-mailboxes
	readObject(ctx, conn, "managed-mailboxes", connectors.Fields("id", "email", "firstName", "lastName", "status"))

	// 10. placement-tests
	readObject(ctx, conn, "placement-tests", connectors.Fields("id", "name", "status", "score"))

	// 11. auto/identitys
	readObject(ctx, conn, "auto/identitys", connectors.Fields("id", "username", "type", "status"))

	// 12. conversation/latest-by-profile (nested responseKey: data.conversations)
	readObject(ctx, conn, "conversation/latest-by-profile", connectors.Fields("id", "title", "is_unread", "platform_type"))

	os.Exit(0)
}

// readObject reads data from a SuperSend object and prints the results.
func readObject(ctx context.Context, conn connectors.ReadConnector, objectName string, fields datautils.Set[string]) {
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     fields,
	})
	if err != nil {
		slog.Error("error reading "+objectName, "error", err)
		return
	}

	slog.Info(objectName, "rows", result.Rows, "done", result.Done)
	printData(result.Data)
}

// printData prints the fields from each record in a readable format.
func printData(data []common.ReadResultRow) {
	for i, row := range data {
		fieldsJSON, _ := json.MarshalIndent(row.Fields, "  ", "  ")
		fmt.Printf("  Record %d Fields:\n  %s\n", i+1, string(fieldsJSON))

		// Also print raw data if fields are empty (for debugging)
		if len(row.Fields) == 0 && len(row.Raw) > 0 {
			rawJSON, _ := json.MarshalIndent(row.Raw, "  ", "  ")
			fmt.Printf("  Record %d Raw:\n  %s\n", i+1, string(rawJSON))
		}
	}
}
