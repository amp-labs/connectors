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
		Since:      time.Now().Add(-30 * time.Hour),
		Fields:     connectors.Fields("Assistant", "Created_By", "Full_Name", "id", "Created_Time"),
		// NextPage:   "https://www.zohoapis.com/crm/v6/Contacts?fields=Assistant%2CCreated_By%2CFull_Name%2Cid%2CCreated_Time\u0026page_token=089df74ef7734aa9f877fa670550bcbafc9c43567bb2f2e2404aa4d2a466a0b2c9432951d4eb3ffe73094c25e18b4c6290eedc53160535ab40b8ed204dd80e7247a90a4d5cd8d69a348dcaeefbccd8087f658d4a72cfa6aaab8ae8e7065246bf6d1fffce6a3eb2a06bab02e3ae935bc3fb63067b3e0da43e421e36b71b7c1bb843d3af99c7679b53100a8f7b8343f012\u0026since=2024-10-22T16%3A15%3A55%2B03%3A00",
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
