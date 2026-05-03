package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	clioTest "github.com/amp-labs/connectors/test/clio"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := clioTest.GetClioGrowConnector(ctx)
	// testscenario.ReadThroughPages(ctx, conn.Grow, common.ReadParams{
	// 	ObjectName: "users",
	// 	Fields:     connectors.Fields("id", "first_name", "last_name", "email"),
	// })
	testscenario.ReadThroughPages(ctx, conn.Grow, common.ReadParams{
		ObjectName: "matters",
		Fields:     connectors.Fields("id", "description", "status", "type"),
	})
	testscenario.ReadThroughPages(ctx, conn.Grow, common.ReadParams{
		ObjectName: "inbox_leads",
		Fields:     connectors.Fields("id", "first_name", "last_name", "state"),
	})
	testscenario.ReadThroughPages(ctx, conn.Grow, common.ReadParams{
		ObjectName: "custom_actions",
		Fields:     connectors.Fields("id", "label", "ui_reference"),
	})
	testscenario.ReadThroughPages(ctx, conn.Grow, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("id", "first_name", "last_name", "emails", "type"),
	})

}
