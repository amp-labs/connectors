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

	conn := connTest.GetNetsuiteRESTAPIConnector(ctx)

	ctx, done = context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "customer",
		Fields:     connectors.Fields("id", "companyName", "email", "phone", "entityStatus"),
	})
	if err != nil {
		utils.Fail("error reading customers from NetSuite REST API", "error", err)
	}

	fmt.Println("Reading customers..")
	utils.DumpJSON(res, os.Stdout)
}
