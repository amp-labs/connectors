package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	connTest "github.com/amp-labs/connectors/test/keap"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "contacts"

// we want to compare fields returned by read and schema properties provided by metadata methods
// they must match for all such objects
func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetKeapConnector(ctx)
	defer utils.Close(conn)

	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for Keap", "error", err)
	}

	fmt.Println("Contacts metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}
