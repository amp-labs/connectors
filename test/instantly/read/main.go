package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/instantly"
	connTest "github.com/amp-labs/connectors/test/instantly"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "campaigns" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetInstantlyConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"name",
		),
	})
	if err != nil {
		utils.Fail("error reading from Instantly", "error", err)
	}

	slog.Info("Reading campaigns..")
	utils.DumpJSON(res, os.Stdout)

	if res.Rows > instantly.DefaultPageSize {
		utils.Fail(fmt.Sprintf("expected max %v rows", instantly.DefaultPageSize))
	}
}
