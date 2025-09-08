package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/highlevelwhitelabel"
	"github.com/amp-labs/connectors/test/highlevelwhitelabel"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := highlevelwhitelabel.GetHighLevelWhiteLabelConnector(context.Background())

	slog.Info("Deleting the business")

	err := testDelete(context.Background(), conn, "businesses", "68921533fe980be6f4421cf8")
	if err != nil {
		return 1
	}

	slog.Info("Deleting the calendars groups")

	err = testDelete(context.Background(), conn, "calendars/groups", "c5d87HDX906XNUdQD3rS")
	if err != nil {
		return 1
	}

	return 0
}

func testDelete(ctx context.Context, conn *ap.Connector, objName string, recordId string) error {
	params := common.DeleteParams{
		ObjectName: objName,
		RecordId:   recordId,
	}

	res, err := conn.Delete(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", objName, err)
	}

	// Print the results.
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
