package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/sellsy"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSellsyConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("first_name", "last_name"),
		// Since:      time.Now().Add(-1 * time.Minute * 12),
		// Until:      time.Now().Add(-1 * time.Minute * 9),
		NextPage: "https://api.sellsy.com/v2/contacts/search?limit=36&offset=WyIzOCJd",
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	slog.Info("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
