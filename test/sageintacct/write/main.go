package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/sageintacct"
	"github.com/amp-labs/connectors/test/sageintacct"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := sageintacct.GetSageIntacctConnector(ctx)

	accountId, err := testCreatingAccount(ctx, conn)
	if err != nil {
		return err
	}

	err = updateAccount(ctx, conn, accountId)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAccount(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "account",
		RecordData: map[string]any{
			"id":                    "15677",
			"name":                  "Vehicle parts - Transmission",
			"accountType":           "balanceSheet",
			"closingType":           "nonClosingAccount",
			"normalBalance":         "debit",
			"alternativeGLAccount":  "none",
			"status":                "active",
			"isTaxable":             false,
			"disallowDirectPosting": true,
		},
	}

	slog.Info("Creating an account...")

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

	return res.RecordId, nil
}

func updateAccount(ctx context.Context, conn *cc.Connector, accountId string) error {
	params := common.WriteParams{
		ObjectName: "account",
		RecordId:   accountId,
		RecordData: map[string]any{
			"name": "Updated Account Name",
		},
	}

	slog.Info("Updating the account...")

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
