package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/gong"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGongConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "transcripts",
		Fields:     connectors.Fields("callId"),
	})
	if err != nil {
		utils.Fail("error reading from Gong", "error", err)
	}

	slog.Info("Reading transcripts..")
	utils.DumpJSON(res, os.Stdout)
}
