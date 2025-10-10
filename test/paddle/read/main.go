package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/paddle"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := paddle.GetPaddleConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id", "name", "email", "status"),
	})
	if err != nil {
		utils.Fail("error reading customers from Paddle", "error", err)
	}

	slog.Info("Reading customers..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "products",
		Fields:     connectors.Fields("id", "name", "status", "description"),
		Since:      time.Now().Add(-2 * time.Minute),
	})
	if err != nil {
		utils.Fail("error reading products from Paddle", "error", err)
	}

	slog.Info("Reading products..")
	utils.DumpJSON(res, os.Stdout)
}
