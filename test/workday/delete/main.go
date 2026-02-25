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

	result, err := conn.Delete(ctx, connectors.DeleteParams{
		ObjectName: "workers",
		RecordId:   "3aa5550b7fe348b98d7b5741afc65534",
	})
	if err != nil {
		utils.Fail("error deleting from workday", "error", err)
	}

	fmt.Println("workday delete result...")
	utils.DumpJSON(result, os.Stdout)
}
