package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/capsule"
	"github.com/amp-labs/connectors/test/utils"
)

var objectName = "parties"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetCapsuleConnector(ctx)

	metadata, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     datautils.NewSet("firstName", "name"),
	})
	if err != nil {
		utils.Fail("error reading data", "error", err)
	}

	fmt.Println("Reading...")
	utils.DumpJSON(metadata, os.Stdout)
}
