package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zohocrm"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx)

	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "users",
		Since:      time.Now().Add(-3000 * time.Hour),
		Fields:     connectors.Fields("Name", "id"),
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
