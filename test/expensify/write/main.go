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

	err = testCreatingReport(ctx, conn)
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

func testCreatingReport(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "report",
		RecordData: map[string]any{
			"type":     "report",
			"policyID": "94C31405ED1893F1",
			"report": map[string]any{
				"title": "Name of the report",
			},
			"employeeEmail": "dipu@withampersand.com",
			"expenses": []map[string]any{
				{
					"date":     "2026-01-01",
					"currency": "USD",
					"merchant": "Name of merchant",
					"amount":   1234,
				},
			},
		},
	}

	slog.Info("Creating new report...")

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
