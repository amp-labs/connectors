package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/goto"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGoToConnector(ctx, providers.ModuleGoTo)

	m, err := conn.ListObjectMetadata(ctx, []string{"historicalMeetings", "webinars"})
	if err != nil {
		utils.Fail("error listing metadata for GoTo", "error", err)
	}

	utils.DumpJSON(m.Result, os.Stdout)
	fmt.Println("Errors: ", m.Errors)
}
