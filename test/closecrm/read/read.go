package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers/closecrm"
	testConn "github.com/amp-labs/connectors/test/closecrm"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := testConn.GetCloseConnector(ctx)

	if err := readActivities(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readContacts(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := readLeads(ctx, conn); err != nil {
		slog.Error(err.Error())
	}
}

func readActivities(ctx context.Context, conn *closecrm.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "activity",
		Since:      time.Now().Add(-72 * time.Hour),
		Fields:     connectors.Fields("user_id", "user_name", "source", "id"),
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

	fmt.Println(string(jsonStr))

	return nil
}

func readContacts(ctx context.Context, conn *closecrm.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "contact",
		Fields:     connectors.Fields("name", "title", "emails", "phones", "id"),
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

	fmt.Println(string(jsonStr))

	return nil
}

func readLeads(ctx context.Context, conn *closecrm.Connector) error {
	config := connectors.ReadParams{
		ObjectName: "lead",
		Since:      time.Now().Add(-72 * time.Hour),
		Fields:     connectors.Fields("display_name", "description", "name", "id"),
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

	fmt.Println(string(jsonStr))

	return nil
}
