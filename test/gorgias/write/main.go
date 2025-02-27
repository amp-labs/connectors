package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	gor "github.com/amp-labs/connectors/providers/gorgias"
	"github.com/amp-labs/connectors/test/gorgias"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := gorgias.GetConnector(ctx)

	err := testCreatingTicket(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingCustomFields(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingRules(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingTags(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingTicket(ctx context.Context, conn *gor.Connector) error {
	params := common.WriteParams{
		ObjectName: "tickets",
		RecordData: map[string]any{
			"messages": []map[string]any{
				{
					"channel": "api",
					"receiver": map[string]any{
						"name": "mimi",
					},
					"sender": map[string]any{
						"name":    "Karage",
						"address": "LT Strees",
					},
					"from_agent": true,
					"public":     true,
					"source": map[string]any{
						"type": "aircall",
						"from": map[string]any{
							"name":    "Karage",
							"address": "+1 457645242",
						},
					},
					"last_sending_error": map[string]any{
						"error": "none",
					},
					"via": "api",
				},
			},
			"priority":   "normal",
			"spam":       false,
			"from_agent": true,
			"status":     "open",
			"via":        "api",
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

func testCreatingCustomFields(ctx context.Context, conn *gor.Connector) error {
	params := common.WriteParams{
		ObjectName: "custom-fields",
		RecordData: map[string]any{
			"object_type": "Ticket",
			"required":    false,
			"label":       "Test-field",
			"definition": map[string]any{
				"data_type": "text",
				"input_settings": map[string]any{
					"input_type": "input",
				},
			},
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

func testCreatingRules(ctx context.Context, conn *gor.Connector) error {
	params := common.WriteParams{
		ObjectName: "rules",
		RecordData: map[string]any{
			"code": "if (eq(ticket.from_agent, false) && eq(ticket.status, 'open')) {if (containsAny(message.intents.name, ['other/no_reply']) || containsAny(ticket.customer.email, ['noreply@','@noreply'])) {Action('addTags', { tags: 'auto-close' });Action('setStatus', { status: 'closed' })}}",
			"name": "Auto-close-all-ticket",
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

func testUpdatingTags(ctx context.Context, conn *gor.Connector) error {
	params := common.WriteParams{
		ObjectName: "tags",
		RecordId:   "34799",
		RecordData: map[string]any{
			"name": "Urgent",
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
