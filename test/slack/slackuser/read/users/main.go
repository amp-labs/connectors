package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	slackshared "github.com/amp-labs/connectors/test/slack"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := slackshared.NewConnector(ctx, providers.SlackUserScope)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name", "real_name", "is_bot"),
	})
}
