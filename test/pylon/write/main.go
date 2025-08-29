package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/pylon"
	"github.com/amp-labs/connectors/test/pylon"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	conn := pylon.GetConnector(ctx)

	_, err := testCreatingTasks(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingTasks(ctx context.Context, conn *cc.Connector) (string, error) {
	params := common.WriteParams{
		ObjectName: "tasks",
		RecordData: map[string]any{
			"title": "Ampersand write test",
		},
	}

	slog.Info("Creating tasks...")

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
