package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	fr "github.com/amp-labs/connectors/providers/front"
	"github.com/amp-labs/connectors/test/front"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := front.GetFrontConnector(ctx)

	err := testCreatingAccount(ctx, conn)
	if err != nil {
		return err
	}

	err = testPatchingAccount(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingLinks(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAccount(ctx context.Context, conn *fr.Connector) error {
	params := common.WriteParams{
		ObjectName: "accounts",
		RecordData: map[string]any{
			"name":        "Test Users",
			"description": "A test user account creation",
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

func testPatchingAccount(ctx context.Context, conn *fr.Connector) error {
	params := common.WriteParams{
		ObjectName: "accounts",
		RecordId:   "acc_fkb9ek",
		RecordData: map[string]any{
			"name": "Update Test Users Account",
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

func testCreatingLinks(ctx context.Context, conn *fr.Connector) error {
	params := common.WriteParams{
		ObjectName: "links",
		RecordData: map[string]any{
			"name":         "A test domain link",
			"external_url": "https://test.com",
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
