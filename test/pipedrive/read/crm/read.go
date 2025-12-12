package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
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

	conn := testConn.GetPipedriveConnector(ctx, providers.ModulePipedriveCRM)

	if err := readActivities(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readDeals(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readStages(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readActivities(ctx context.Context, conn *pipedrive.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "activities",
		Since:      time.Now().Add(-720 * time.Hour),
		Until:      time.Now().Add(-1 * time.Hour),
		Fields:     connectors.Fields("owner_id", "done", "id"),
		// NextPage:   "https://api.pipedrive.com/api/v2/activities?cursor=eyJmaWVsZCI6ImlkIiwiZmllbGRWYWx1ZSI6MTMsInNvcnREaXJlY3Rpb24iOiJhc2MiLCJpZCI6MTN9\u0026limit=2\u0026updated_since=2025-09-27T18%3A42%3A31%2B03%3A00",
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
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

func readDeals(ctx context.Context, conn *pipedrive.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "deals",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("close_time", "id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
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

func readStages(ctx context.Context, conn *pipedrive.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "stages",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("id"),
	}

	result, err := conn.Read(ctx, config)
	if err != nil {
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
