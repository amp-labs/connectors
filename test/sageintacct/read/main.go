package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/sageintacct"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSageIntacctConnector(ctx)

	slog.Info("Testing account read...")
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "contact",
		Fields:     connectors.Fields("id", "key", "href"),
	})
	if err != nil {
		utils.Fail("error reading accounts from Sage Intacct", "error", err)
	}

	slog.Info("Reading accounts..")
	utils.DumpJSON(res, os.Stdout)
}
