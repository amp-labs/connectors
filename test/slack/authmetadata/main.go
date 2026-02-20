package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	slackConn "github.com/amp-labs/connectors/providers/slack"
	"github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := slack.NewConnector(ctx)

	info, err := conn.GetPostAuthInfo(ctx)
	if err != nil || info.CatalogVars == nil {
		utils.Fail("error obtaining auth info", "error", err)
	}

	teamId := slackConn.NewAuthMetadataVars(*info.CatalogVars).TeamId

	// Log the retrieved team ID.
	slog.Info("retrieved auth metadata", "team id", teamId)
}
