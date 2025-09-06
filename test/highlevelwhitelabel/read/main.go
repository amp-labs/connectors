package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/highlevelwhitelabel"
	"github.com/amp-labs/connectors/test/highlevelwhitelabel"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := highlevelwhitelabel.GetHighLevelWhiteLabelConnector(context.Background())

	err := testRead(context.Background(), conn, "businesses", []string{""})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "products/collections", []string{""})
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "calendars/groups", []string{""})
	if err != nil {
		return 1
	}

	return 0
}

func testRead(ctx context.Context, conn *ap.Connector, objName string, fields []string) error {
	params := common.ReadParams{
		ObjectName: objName,
		Fields:     connectors.Fields(fields...),
	}

	res, err := conn.Read(ctx, params)
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
