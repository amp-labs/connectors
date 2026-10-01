package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetJiraConnector(ctx)

	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	cloudId := (*info.CatalogVars)["cloudId"]

	if len(cloudId) == 0 {
		utils.Fail("missing cloud id in post authentication metadata")
	}

	slog.Info("retrieved auth metadata", "cloud id", cloudId)
}
