package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/odoo"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "crm.lead",
		Fields:     connectors.Fields("id", "name", "write_date"),
	})
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "res.partner",
		Fields:     connectors.Fields("id", "name", "write_date"),
	})

}
