package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/quickbooks"
	"github.com/amp-labs/connectors/test/quickbooks"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := quickbooks.GetQuickBooksConnector(ctx)

	_, err := testCreatingAccount(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAccount(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"Name":        "MyJobs_test",
			"AccountType": "Accounts Receivable",
		},
	}

	slog.Info("Creating account...")

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return res.Data["Id"].(string), nil
}
