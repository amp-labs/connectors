package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx, providers.ModuleZohoServiceDeskPlus)

	if err := createTasks(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createChanges(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateChanges(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func createTasks(ctx context.Context, conn *zoho.Connector) error {
	config := common.WriteParams{
		ObjectName: "tasks",
		RecordData: map[string]any{
			"percentage_completion":  "30",
			"estimated_effort_hours": "20",
			"email_before":           "3600000",
			"description":            "The SRS must contain all the requirements for the feature",
			"title":                  "Create SRS",
			"additional_cost":        "100",
			"actual_end_time": map[string]any{
				"value": "1512974940000",
			},
			"actual_start_time": map[string]any{
				"value": "1421988300000",
			},
			"scheduled_end_time": map[string]any{
				"value": "1512974940000",
			},
			"estimated_effort_minutes": "45",
			"estimated_effort_days":    "15",
			"scheduled_start_time": map[string]any{
				"value": "1421988300000",
			},
		},
	}
	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createChanges(ctx context.Context, conn *zoho.Connector) error {
	config := common.WriteParams{
		ObjectName: "changes",
		RecordData: map[string]any{
			"description": "test-description",
			"title":       "test-title",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func updateChanges(ctx context.Context, conn *zoho.Connector) error {
	config := common.WriteParams{
		ObjectName: "changes",
		RecordId:   "239248000000370134",
		RecordData: map[string]any{
			"description": "tests",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
