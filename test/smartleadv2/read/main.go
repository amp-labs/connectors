package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/smartleadv2"
	"github.com/amp-labs/connectors/test/utils"
)

var objectNames = []string{"campaigns", "client", "email-accounts"} // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSmartleadV2Connector(ctx)

	for _, objectName := range objectNames {
		res, err := conn.Read(ctx, common.ReadParams{
			ObjectName: objectName,
			Fields:     connectors.Fields("name"),
		})
		if err != nil {
			utils.Fail("error reading from Smartlead", "error", err)
		}

		fmt.Printf("Reading %s...\n", objectName)
		utils.DumpJSON(res, os.Stdout)
	}
}
