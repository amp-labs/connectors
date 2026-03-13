package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	co "github.com/amp-labs/connectors/providers/chargeover"
	"github.com/amp-labs/connectors/test/chargeover"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := chargeover.NewConnector(ctx)

	err := testCreatingCustomer(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateCustomer(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingCustomer(ctx context.Context, conn *co.Connector) error {
	params := common.WriteParams{
		ObjectName: "customer",
		RecordData: map[string]any{
			"company": "Another one",
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

func testUpdateCustomer(ctx context.Context, conn *co.Connector) error {
	params := common.WriteParams{
		ObjectName: "customer",
		RecordId:   "2",
		RecordData: map[string]any{
			"campaign_details": "campaign ...",
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
