package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/netsuite"
	"github.com/amp-labs/connectors/test/utils"
)

const TimeoutSeconds = 180

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNetsuiteSuiteQLConnector(ctx)

	ctx, done = context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "account",
		Fields:     connectors.Fields("id", "acctname", "acctnumber"),
		Since:      time.Now().Add(-1 * time.Hour),
	})
	if err != nil {
		utils.Fail("error reading accounts from NetSuite SuiteQL", "error", err)
	}

	fmt.Println("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)
}
