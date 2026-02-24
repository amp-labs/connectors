package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/workday"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetWorkdayConnector(ctx)

	result, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "workers",
		Fields:     connectors.Fields("id", "descriptor", "primaryWorkEmail"),
	})
	if err != nil {
		utils.Fail("error reading from workday", "error", err)
	}

	fmt.Println("workday read result...")
	utils.DumpJSON(result, os.Stdout)
}
