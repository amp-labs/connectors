package main

import (
	"context"
	"fmt"
	"os"
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

	res, err := conn.GetRecordsWithIds(ctx,
		"Account",
		[]string{
			"001ak00000OQ4RxAAL",
			"001ak00000OQ4RyAAL",
			"001ak00000OQ4TZAA1",
		},
		[]string{"id", "name", "shippingstreet"},
		nil)
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading..")
	utils.DumpJSON(res, os.Stdout)
}
