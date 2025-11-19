package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cc "github.com/amp-labs/connectors/providers/solarwinds"
	"github.com/amp-labs/connectors/test/solarwinds"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()

	// Set up slog logging.
	utils.SetupLogging()

	conn := solarwinds.GetSolarWindsConnector(ctx)

	err := testCreatingIncident(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingIncident(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateIncident(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingIncident(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "incidents",
		RecordData: map[string]any{
			"name": "Payment Gateway down",
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

func testUpdateIncident(ctx context.Context, conn *cc.Connector) error {
	params := common.WriteParams{
		ObjectName: "incidents",
		RecordId:   "169255271",
		RecordData: map[string]any{
			"name": "CICD Pipeline broken",
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
