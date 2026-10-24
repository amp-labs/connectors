package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/circleback"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "meetings",
		Fields:     connectors.Fields("id", "name", "createdAt", "updatedAt", "attendees", "notes", "transcript"),
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "action-items",
		Fields:     connectors.Fields("id", "title", "status", "meetingIds"),
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "tag",
		Fields:     connectors.Fields("id", "name"),
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "people",
		Fields:     connectors.Fields("id", "email", "companyId"),
		PageSize:   10,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "companies",
		Fields:     connectors.Fields("name", "domain"),
	})
}
