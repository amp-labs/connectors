package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	fd "github.com/amp-labs/connectors/providers/freshdesk"
	"github.com/amp-labs/connectors/test/freshdesk"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}
}

func run() error {
	ctx := context.Background()
	conn := freshdesk.GetFreshdeskConnector(ctx)

	err := testCreatingTicket(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateTicket(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingContacts(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingTicket(ctx context.Context, conn *fd.Connector) error {
	params := common.WriteParams{
		ObjectName: "tickets",
		RecordData: map[string]any{
			"description": "Details about the issue...",
			"subject":     "Support Needed...",
			"email":       "tom@outerspace.com",
			"priority":    1,
			"status":      2,
			"cc_emails":   []string{"ram@freshdesk.com", "diana@freshdesk.com"},
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

func testUpdateTicket(ctx context.Context, conn *fd.Connector) error {
	params := common.WriteParams{
		ObjectName: "tickets",
		RecordId:   "3",
		RecordData: map[string]any{
			"subject": "Support Needed... ASAP",
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

func testCreatingContacts(ctx context.Context, conn *fd.Connector) error {
	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"name":         "Super Man",
			"email":        gofakeit.Email(),
			"other_emails": []string{gofakeit.Email()},
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
