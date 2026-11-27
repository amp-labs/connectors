package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/twenty"
	"github.com/amp-labs/connectors/test/utils"
)

// Describes the objects named on the command line, e.g. a custom object: go run ./test/twenty/metadata vendorContracts
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetTwentyConnector(ctx)

	objectNames := os.Args[1:]
	if len(objectNames) == 0 {
		objectNames = []string{"companies", "people"}
	}

	res, err := conn.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		utils.Fail("error listing metadata for Twenty", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
