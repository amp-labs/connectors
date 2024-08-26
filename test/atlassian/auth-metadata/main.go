package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/atlassian"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAtlassianConnector(ctx)
	defer utils.Close(conn)

	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	metadata := atlassian.NewAuthMetadataVars(*info.CatalogVars)
	cloudId := metadata.CloudId

	if len(cloudId) == 0 {
		utils.Fail("missing cloud id in post authentication metadata")
	}

	slog.Info("retrieved auth metadata", "cloud id", cloudId)
}
