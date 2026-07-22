package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/jump"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "meetings",
		Fields:     connectors.Fields("id", "status", "source", "startedAt"),
		PageSize:   2,
	})
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "scorecards",
		Fields:     connectors.Fields("id", "name", "criteria"),
		PageSize:   2,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "tasks",
		Fields:     connectors.Fields("id", "name", "assignee"),
		PageSize:   2,
	})
}
