package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/fathom"
	"github.com/amp-labs/connectors/test/fathom"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := fathom.GetFathomConnector(ctx)

	_, err := testCreatingWebhooks(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingWebhooks(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "webhooks",
		RecordData: map[string]any{
			"destination_url":      "https://play.svix.com/in/e_5U95s0OihUbc32B8UDA1MoAaAG2/",
			"include_action_items": true,
		},
	}

	slog.Info("Creating webhooks...")

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
