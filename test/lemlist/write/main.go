package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ll "github.com/amp-labs/connectors/providers/lemlist"
	"github.com/amp-labs/connectors/test/lemlist"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := lemlist.GetLemlistConnector(ctx)

	err := testCreatingCampaigns(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingWebhooks(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCampaigns(ctx context.Context, conn *ll.Connector) error {
	params := common.WriteParams{
		ObjectName: "campaigns",
		RecordData: map[string]any{
			"name": "Eclipse of Empires",
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

func testCreatingWebhooks(ctx context.Context, conn *ll.Connector) error {
	params := common.WriteParams{
		ObjectName: "hooks",
		RecordData: map[string]any{
			"targetUrl": "https://webhook.site/f748c494-e583-4b47-b9ea",
			"type":      "emailsOpened",
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
