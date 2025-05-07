package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	ap "github.com/amp-labs/connectors/providers/instantlyai"
	"github.com/amp-labs/connectors/test/instantlyai"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	conn := instantlyai.GetInstantlyAIConnector(context.Background())

	err := testRead(context.Background(), conn, "api-keys")
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "custom-tags")
	if err != nil {
		return 1
	}

	err = testRead(context.Background(), conn, "lead-lists")
	if err != nil {
		return 1
	}

	return 0
}

func testRead(ctx context.Context, conn *ap.Connector, objName string) error {
	params := common.ReadParams{
		ObjectName: objName,
		Fields:     connectors.Fields(""),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to read campaign: %w", err)
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
