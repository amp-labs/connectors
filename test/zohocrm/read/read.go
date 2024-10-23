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
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zohocrm"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := zohocrm.GetZohoConnector(ctx)
	defer utils.Close(conn)

	if err := readContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readDeals(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readLeads(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

}

func readContacts(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "contacts",
		Since:      time.Now().Add(-10 * time.Hour),
		Fields:     connectors.Fields("Assistant", "Created_By", "Full_Name", "id", "Created_Time"),
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

func readDeals(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "deals",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("Account_Name", "Closing_Date", "id"),
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

func readLeads(ctx context.Context, conn connectors.ReadConnector) error {
	config := connectors.ReadParams{
		ObjectName: "leads",
		Since:      time.Now().Add(-720 * time.Hour),
		Fields:     connectors.Fields("Converted_Date_Time", "Email", "Record_Status__s", "id"),
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
