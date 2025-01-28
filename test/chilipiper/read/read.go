package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/chilipiper"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := chilipiper.GetChiliPiperConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "workspace",
		Fields:     connectors.Fields("name"),
		// Read newsletters using pagination, where it omits first record.
	})
	if err != nil {
		utils.Fail("error reading from ChiliPiper App", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
