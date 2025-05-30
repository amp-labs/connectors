package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	dx "github.com/amp-labs/connectors/providers/dixa"
	"github.com/amp-labs/connectors/test/dixa"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := dixa.GetConnector(ctx)

	err := testCreatingAgent(ctx, conn)
	if err != nil {
		return err
	}

	err = patchAgent(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingTeams(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingQueue(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAgent(ctx context.Context, conn *dx.Connector) error {
	params := common.WriteParams{
		ObjectName: "agents",
		RecordData: map[string]any{
			"displayName": "Alice Brown",
			"email":       gofakeit.Email(),
			"phoneNumber": "+1" + gofakeit.Phone(),
			"firstName":   "Alice",
			"lastName":    "Brown",
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

func patchAgent(ctx context.Context, conn *dx.Connector) error {
	params := common.WriteParams{
		ObjectName: "agents",
		RecordId:   "af34671b-f191-4ecf-884f-6e28abe82b39",
		RecordData: map[string]any{
			"firstName": "Charles",
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

func testCreatingTeams(ctx context.Context, conn *dx.Connector) error {
	params := common.WriteParams{
		ObjectName: "teams",
		RecordData: map[string]any{
			"name": "Integration Team",
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

func testCreatingQueue(ctx context.Context, conn *dx.Connector) error {
	params := common.WriteParams{
		ObjectName: "queues",
		RecordData: map[string]any{
			"name":              "MyCorp main queue",
			"callFunctionality": false,
			"isDefault":         true,
			"queueThresholds": map[string]any{
				"WaitingConversations": 10,
			},
			"offerTimeout":                10,
			"offerAlgorithm":              "AllAtOnce",
			"wrapupTimeout":               10,
			"priority":                    1,
			"offerAbandonedConversations": false,
			"doNotOfferTimeouts": map[string]any{
				"WidgetChat": 1,
			},
			"isDoNotOfferEnabled": true,
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
