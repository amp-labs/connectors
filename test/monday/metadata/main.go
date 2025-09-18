package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/monday"
	"github.com/amp-labs/connectors/test/utils"
)

var objects = []string{"boards", "users"} // nolint: gochecknoglobals

// We want to compare fields returned by read and schema properties provided by metadata methods.
// Properties from read must all be present in schema definition.
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMondayConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		utils.Fail("error listing metadata for Monday", "error", err)
	}

	utils.DumpJSON(m, os.Stdout)
}
