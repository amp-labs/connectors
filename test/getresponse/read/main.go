package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetTheGetResponseConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("id", "name", "description"),
	})
	if err != nil {
		utils.Fail("error reading from GetResponse", "error", err)
	}

	fmt.Println("Reading campaigns..")
	utils.DumpJSON(res, os.Stdout)
}
