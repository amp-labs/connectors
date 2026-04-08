package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/test/slack"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := slack.NewConnector(ctx)

	callId, err := testCreatingCalls(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingCalls(ctx, conn, callId)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCalls(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "calls",
		RecordData: map[string]any{
			"join_url":           "https://example.com/join",
			"external_unique_id": fmt.Sprintf("ext-id-%d", os.Getpid()),
		},
	}

	slog.Info("Creating calls...")

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

	return res.Data["id"].(string), nil
}

func testUpdatingCalls(ctx context.Context, conn *cc.Connector, callId string) error {
	params := common.WriteParams{
		ObjectName: "calls",
		RecordId:   callId,
		RecordData: map[string]any{
			"title": "Updated Call Name",
		},
	}

	slog.Info("Updating calls...")

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
