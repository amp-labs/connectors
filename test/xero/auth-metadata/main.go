package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/xero"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/xero"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetXeroConnector(ctx)

	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	tenantId := xero.NewAuthMetadataVars(*info.CatalogVars).TenantId

	// Log the retrieved tenant ID.
	slog.Info("retrieved auth metadata", "tenant id", tenantId)
}
