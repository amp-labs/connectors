package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	cl "github.com/amp-labs/connectors/providers/callrail"
	"github.com/amp-labs/connectors/test/callrail"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := callrail.NewConnector(ctx)

	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	accountID := cl.NewAuthMetadataVars(*info.CatalogVars).AccountID

	// Log the retrieved tenant ID.
	slog.Info("retrieved auth metadata", "accountId", accountID)
}
