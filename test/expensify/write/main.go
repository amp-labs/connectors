package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/expensify"
	"github.com/amp-labs/connectors/test/expensify"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := expensify.GetConnector(ctx)

	err := testCreatingPolicty(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingPolicty(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "policy",
		RecordData: map[string]any{
			"type":       "policy",
			"policyName": "My New Policy",
		},
	}

	slog.Info("Creating new policy...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
