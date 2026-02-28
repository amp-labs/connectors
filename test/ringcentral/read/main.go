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
	rc "github.com/amp-labs/connectors/providers/ringcentral"
	"github.com/amp-labs/connectors/test/ringcentral"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn, err := ringcentral.NewConnector(ctx)
	if err != nil {
		utils.Fail("error creating ringcentral connector", "error", err)
	}

	if err := readContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readMeeting(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readLeads(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readCallRecordings(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readContacts(ctx context.Context, conn *rc.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "contacts",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("uri", "firstName", "lastName", "id"),
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

func readMeeting(ctx context.Context, conn *rc.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "meetings",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("startTime", "hostInfo", "displayName"),
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

func readLeads(ctx context.Context, conn *rc.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "comm-handling/states",
		// Since:      time.Now().Add(-720 * time.Hour),
		Fields: connectors.Fields("enabled", "displayName", "conditions"),
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

func readCallRecordings(ctx context.Context, conn *rc.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "call-log",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("uri", "sessionId", "type"),
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
