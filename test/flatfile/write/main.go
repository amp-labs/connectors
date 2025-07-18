package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/flatfile"
	"github.com/amp-labs/connectors/test/flatfile"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := flatfile.GetConnector(ctx)

	appId, err := testCreatingApps(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateApps(ctx, conn, appId)
	if err != nil {
		return err
	}

	err = testCreateSpaces(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingApps(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "apps",
		RecordData: map[string]any{
			"name":         "Nightly Data Loads",
			"namespace":    "nightly-data",
			"type":         "CUSTOM",
			"entity":       "Sync",
			"entityPlural": "Syncs",
			"icon":         "<svg xmlns=\"http://www.w3.org/2000/svg\" fill=\"currentColor\" class=\"bi bi-bar-chart-fill\" viewBox=\"0 0 16 16\">\n  <path d=\"M1 11a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1zm5-4a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v7a1 1 0 0 1-1 1H7a1 1 0 0 1-1-1zm5-5a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1h-2a1 1 0 0 1-1-1z\"/>\n</svg>",
			"metadata": map[string]any{
				"foo": "bar",
			},
		},
	}

	slog.Info("Creating an app...")
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

func testUpdateApps(ctx context.Context, conn *cc.Connector, appId string) error {
	params := common.WriteParams{
		ObjectName: "apps",
		RecordId:   appId,
		RecordData: map[string]any{
			"name": "Updated Nightly Data Loads",
		},
	}

	slog.Info("Updating app...")
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

func testCreateSpaces(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "spaces",
		RecordData: map[string]any{
			"displayOrder":      1,
			"environmentId":     "us_env_cEqqBV0z",
			"name":              "My First Workbook",
			"primaryWorkbookId": "us_wb_YOUR_ID",
		},
	}

	slog.Info("Creating a space...")
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
