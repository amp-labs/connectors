package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/notion"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetNotionConnector(ctx)

	slog.Info("=== Reading users ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("id", "name", "type"),
		PageSize:   2,
	})

	slog.Info("=== Reading file uploads ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "file_uploads",
		Fields:     connectors.Fields("id", "filename", "status", "last_edited_time"),
		PageSize:   2,
	})
}
