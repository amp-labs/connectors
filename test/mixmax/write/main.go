package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	mx "github.com/amp-labs/connectors/providers/mixmax"
	"github.com/amp-labs/connectors/test/mixmax"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := mixmax.GetConnector(ctx)

	err := testCreatingCodeSnippet(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingInsightReports(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingMessages(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCodeSnippet(ctx context.Context, conn *mx.Connector) error {
	params := common.WriteParams{
		ObjectName: "codesnippets",
		RecordData: map[string]any{
			"background": "rgb(255, 255, 255)",
			"theme":      "ambiance",
			"language":   "javascript",
			"html":       "<h1>Test Header</h1>",
		},
	}

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

func testCreatingInsightReports(ctx context.Context, conn *mx.Connector) error {
	params := common.WriteParams{
		ObjectName: "insightsreports",
		RecordData: map[string]any{
			"title": "Super report",
			"type":  "messages",
			"query": "sent:last30days from:everyone",
		},
	}

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

func testCreatingMessages(ctx context.Context, conn *mx.Connector) error {
	params := common.WriteParams{
		ObjectName: "messages",
		RecordData: map[string]any{
			"to": []map[string]any{
				{"email": "alice@example.com", "name": "Alice Example"},
			},
			"cc":      []map[string]any{{"email": "bob@example.com"}},
			"subject": "API Test", "body": "I love Mixmax!.",
			"trackingEnabled":      true,
			"linkTrackingEnabled":  false,
			"notificationsEnabled": false,
			"userHasModified":      false,
		},
	}

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
