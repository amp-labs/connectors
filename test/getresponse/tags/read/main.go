package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/getresponse"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGetResponseConnector(ctx)

	slog.Info("=== Basic paginated read: tags ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "tags",
		Fields:     connectors.Fields("tagId", "name", "color", "createdAt"),
		PageSize:   10,
	})

	slog.Info("Tags read tests completed successfully!")
}
