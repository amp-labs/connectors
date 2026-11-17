package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	conn := connTest.GetGoogleMailConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "messages",
		Fields: connectors.Fields(
			"id",
			"threadId",
			"labelIds",
			"snippet",
			"$['payload']",
			"$['payload']['body']['data']",
		),
		Since:    time.Now().Add(-1 * time.Minute * 2),
		PageSize: 10,
	})
}
