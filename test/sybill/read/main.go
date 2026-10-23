package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/sybill"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	// Transcript and summary make the connector fetch each conversation's detail.
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "conversations",
		Fields:     connectors.Fields("conversationId", "title", "startTime", "crm", "summary", "transcript"),
		Since:      time.Now().Add(-30 * 24 * time.Hour),
		PageSize:   2,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("accountId", "name", "latestDeal"),
		PageSize:   2,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "deals",
		Fields:     connectors.Fields("dealId", "name", "stage", "amount"),
		PageSize:   2,
	})
}
