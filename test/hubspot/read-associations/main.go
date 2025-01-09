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
	"github.com/amp-labs/connectors/common/logging"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Turn on logging
	ctx = logging.WithLoggerEnabled(ctx, true)

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "companies",
		Fields:     connectors.Fields("name"),
		NextPage:   "",
		Since:      time.Now().Add(-24 * 365 * 10 * time.Hour), // 10 years should cover it
		AssociatedObjects: []string{
			"contacts",
		},
	})
	if err != nil {
		utils.Fail("error reading from hubspot", "error", err)
	}

	fmt.Println("Reading contacts..") //nolint:forbidigo

	// Dump the result.
	utils.DumpJSON(res, os.Stdout)
}
