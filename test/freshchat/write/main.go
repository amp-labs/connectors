package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	fc "github.com/amp-labs/connectors/providers/freshchat"
	"github.com/amp-labs/connectors/test/freshchat"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := freshchat.NewConnector(ctx)

	err := testCreatingUsers(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdatingUsers(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingUsers(ctx context.Context, conn *fc.Connector) error {
	params := common.WriteParams{
		ObjectName: "users",
		RecordData: map[string]any{
			"avatar": map[string]any{
				"url": "https://web.freshchat.com/img/johndoe.png",
			},
			"email":        "milton.doe@mail.com",
			"first_name":   "Milton",
			"last_name":    "Doe",
			"phone":        "235689714",
			"reference_id": "milton@doe",
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

func testUpdatingUsers(ctx context.Context, conn *fc.Connector) error {
	params := common.WriteParams{
		ObjectName: "users",
		RecordId:   "11ac8e40-9c32-4e6b-a97b-f4fa9321c916",
		RecordData: map[string]any{
			"avatar": map[string]any{
				"url": "https://web.freshchat.com/img/johndoe2.png",
			},
			"timezone": "Africa/nairobi",
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
