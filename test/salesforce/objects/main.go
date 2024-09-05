package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	defer utils.Close(conn)

	objects, err := conn.GetSupportedObjects(ctx)
	if err != nil {
		utils.Fail("couldn't retrieve supported objects for Salesforce", "error", err)
	}

	for _, o := range objects {
		fmt.Println(o.Name)
	}
}
