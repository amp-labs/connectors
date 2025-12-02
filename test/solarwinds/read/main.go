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
	"github.com/amp-labs/connectors/providers/solarwinds"
	conn "github.com/amp-labs/connectors/test/solarwinds"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := conn.GetSolarWindsConnector(ctx)

	if err := readIncidents(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readDepartmets(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readMobiles(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readIncidents(ctx context.Context, conn *solarwinds.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "incidents",
		Since:      time.Now().Add(-1000 * time.Hour),
		Until:      time.Now().Add(-1 * time.Hour),
		Fields:     connectors.Fields("number", "name", "id"),
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

func readDepartmets(ctx context.Context, conn *solarwinds.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "departments",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("description", "id"),
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

func readMobiles(ctx context.Context, conn *solarwinds.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "mobiles",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("status", "name", "category", "id"),
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
