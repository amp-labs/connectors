package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/snapchatads"
	ap "github.com/amp-labs/connectors/test/snapchatads"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := ap.GetConnector(ctx)

	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	metadata := snapchatads.NewAuthMetadataVars(*info.CatalogVars)
	organizationId := metadata.OrganizationId

	if len(organizationId) == 0 {
		utils.Fail("missing organization id in post authentication metadata")
	}

	slog.Info("retrieved auth metadata", "organization id", organizationId)
}
