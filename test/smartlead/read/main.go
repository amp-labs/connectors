package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/smartlead"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "campaigns" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSmartleadConnector(ctx)
	defer utils.Close(conn)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("name", "status", "user_id"),
	})
	if err != nil {
		utils.Fail("error reading from Smartlead", "error", err)
	}

	fmt.Println("Reading campaign..")
	utils.DumpJSON(res, os.Stdout)
}
