package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPipelinerConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Contacts",
		Fields:     connectors.Fields("id", "formatted_name", "account_position"),
	})
	if err != nil {
		utils.Fail("error reading from Pipeliner", "error", err)
	}

	fmt.Println("Reading Contacts..")
	utils.DumpJSON(res, os.Stdout)
}
