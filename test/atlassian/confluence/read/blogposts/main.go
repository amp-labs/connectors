package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	conn := connTest.GetConfluenceConnector(ctx)

	/**
	4 - "2026-02-25T05:18:12.185Z" -- newest
	3 - "2026-02-25T05:18:01.014Z"
	2 - "2026-02-25T05:17:45.747Z"
	1 - "2026-02-25T05:17:34.013Z" -- oldest
	*/

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "blogposts",
		Fields:     connectors.Fields("title"),
		Since:      utils.Timestamp("2026-02-25T05:18:01.014Z"),
		PageSize:   3,
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(res, os.Stdout)
}
