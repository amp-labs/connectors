package main

import (
	"context"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/constantcontact"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConstantContactConnector(ctx)

	testscenario.ValidateMetadataExactlyMatchesRead(ctx, conn, "privileges")
}
