package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive"
	testConn "github.com/amp-labs/connectors/test/pipedrive"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testConn.GetPipedriveConnector(ctx, providers.ModulePipedriveLegacy)

	if err := createActivity(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateActivity(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createCallLog(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createCallLog(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func createActivity(ctx context.Context, conn *pipedrive.Connector) error {
	config := common.WriteParams{
		ObjectName: "activities",
		RecordData: map[string]any{
			"due_date":           "2024-10-30",
			"location":           "Dar es salaam",
			"public_description": "Demo activity",
			"subject":            "I usually can't come up with words",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		fmt.Println("Object: ", config.ObjectName)
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func updateActivity(ctx context.Context, conn *pipedrive.Connector) error {
	config := common.WriteParams{
		ObjectName: "activities",
		RecordId:   "1",
		RecordData: map[string]any{
			"done": "1",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		fmt.Println("Object: ", config.ObjectName)
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func createCallLog(ctx context.Context, conn *pipedrive.Connector) error {
	config := common.WriteParams{
		ObjectName: "callLogs",
		RecordData: map[string]any{
			"outcome":         "connected",
			"to_phone_number": "+1234567890",
			"subject":         "string",
			"start_time":      time.Now().Add(-4 * time.Hour).Format(time.RFC3339),
			"end_time":        time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			"activity_id":     "1",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		fmt.Println("Object: ", config.ObjectName)
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
