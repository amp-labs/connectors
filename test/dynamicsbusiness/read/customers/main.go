package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/dynamicsbusiness"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetDynamicsBusinessCentralConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Customers",
		Fields:     connectors.Fields("id", "displayName", "email"),
		//NextPage:   "https://api.businesscentral.dynamics.com/v2.0/5c6241d0-74cc-48a2-b667-3eb0d738af72/Production/api/v2.0/companies(70c0c603-f4f9-ef11-9344-6045bdc8c234)/customers?$select=id%2CdisplayName%2Cemail&aid=FIN&$skiptoken=75e545bc-f6f9-ef11-9344-6045bdc8c234", // nolint:lll
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	slog.Info("Reading")
	utils.DumpJSON(res, os.Stdout)
}
