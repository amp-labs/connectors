package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx, providers.ModuleZohoDesk)

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "contacts",
		Since:      time.Now().Add(-2 * time.Hour),
		Fields:     connectors.Fields("id", "firstName", "lastName", "isAnonymous"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading... Contacts")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "tickets",
		Since:      time.Now().Add(-3000 * time.Hour),
		Fields:     connectors.Fields("id", "ticketNumber", "email", "status"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading... Tickets")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "departments",
		Since:      time.Now().Add(-1 * time.Hour),
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading... Calls")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, connectors.ReadParams{
		ObjectName: "articles",
		Since:      time.Now().Add(-1 * time.Hour),
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading... Articles")
	utils.DumpJSON(res, os.Stdout)
}
