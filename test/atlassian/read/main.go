package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/atlassian"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAtlassianConnector(ctx)

	day := 24 * time.Hour

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "issues",
		Fields:     connectors.Fields("id", "summary", "status"),
		// NextPage:   "EAIYwIjPuqEzIhZ1cGRhdGVkID4gIi0yOTM1NTQ2NW0i",
		// Below is the example to get issues that were updated in the last 15 min.
		Since: time.Now().Add(-1000 * day),
		Until: time.Now().Add(-120 * day),
	})
	if err != nil {
		utils.Fail("error reading from Atlassian", "error", err)
	}

	fmt.Println("Reading issue..")
	utils.DumpJSON(res, os.Stdout)
}
