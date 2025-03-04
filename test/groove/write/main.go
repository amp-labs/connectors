package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	gr "github.com/amp-labs/connectors/providers/groove"
	"github.com/amp-labs/connectors/test/groove"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := groove.GetConnector(ctx)

	err := testCreatingTicket(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingWebhooks(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingTicket(ctx context.Context, conn *gr.Connector) error {
	params := common.WriteParams{
		ObjectName: "tickets",
		RecordData: map[string]any{
			"body": "Test Ticket Body",
			"from": "integration.user+groove1@withampersand.com",
			"to":   "josephkarage@gmail.com",
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

func testCreatingWebhooks(ctx context.Context, conn *gr.Connector) error {
	params := common.WriteParams{
		ObjectName: "webhooks",
		RecordData: map[string]any{
			"event": "ticket_started",
			"url":   "https://play.svix.com/in/e_GrcBy3b17pfn3n9U3cPxvmC3ENN/",
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
