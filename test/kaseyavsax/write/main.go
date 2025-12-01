package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	kaseya "github.com/amp-labs/connectors/providers/kaseyavsax"
	"github.com/amp-labs/connectors/test/kaseyavsax"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := kaseyavsax.NewConnector(ctx)

	err := testCreatingOrg(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateOrg(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingOrg(ctx context.Context, conn *kaseya.Connector) error {
	params := common.WriteParams{
		ObjectName: "organizations",
		RecordData: map[string]any{
			"Name": "another Niute",
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

func testUpdateOrg(ctx context.Context, conn *kaseya.Connector) error {
	params := common.WriteParams{
		ObjectName: "organizations",
		RecordId:   "3",
		RecordData: map[string]any{
			"Name": "GitLab",
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
