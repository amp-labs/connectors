package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)

	contactNameSchema, err := conn.GetSchema(ctx, "contacts")
	if err != nil {
		utils.Fail("error getting schema", "error", err)
	}

	jsonStrFromContactName, _ := json.MarshalIndent(contactNameSchema, "", "    ")
	fmt.Println("Schema for contacts: ", string(jsonStrFromContactName))

	contactTypeSchema, err := conn.GetSchema(ctx, "0-1")
	if err != nil {
		utils.Fail("error getting schema", "error", err)
	}

	jsonStrFromContactType, _ := json.MarshalIndent(contactTypeSchema, "", "    ")

	fmt.Println("Schema for 0-1: ", string(jsonStrFromContactType))
}
