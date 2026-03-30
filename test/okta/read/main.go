package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/okta"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetOktaConnector(ctx)

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "status", "profile", "created", "lastUpdated"),
		PageSize:   10,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "groups",
		Fields:     connectors.Fields("id", "type", "profile", "created", "lastUpdated"),
		PageSize:   10,
	})

	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "apps",
		Fields:     connectors.Fields("id", "name", "label", "status", "created"),
		PageSize:   10,
	})
}
