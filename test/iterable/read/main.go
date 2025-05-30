package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/iterable"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetIterableConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "channels",
		Fields:     connectors.Fields("id", "name"),
	})
	if err != nil {
		utils.Fail("error reading from Iterable", "error", err)
	}

	fmt.Println("Reading channels..")
	utils.DumpJSON(res, os.Stdout)
}
