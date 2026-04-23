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
	"github.com/amp-labs/connectors/providers/talkdesk"
	testTalkdesk "github.com/amp-labs/connectors/test/talkdesk"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testTalkdesk.NewConnector(ctx)

	if err := readContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readDNCL(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readRecordLists(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readContacts(ctx context.Context, conn *talkdesk.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "contacts",
		Since:      time.Now().Add(-1000 * time.Hour),
		Until:      time.Now().Add(-1 * time.Hour),
		Fields:     connectors.Fields("phones", "name", "id"),
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

func readDNCL(ctx context.Context, conn *talkdesk.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "do-not-call-lists",
		Since:      time.Now().Add(-7020 * time.Hour),
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

func readRecordLists(ctx context.Context, conn *talkdesk.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "record-lists",
		Since:      time.Now().Add(-100000 * time.Hour),
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
