package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/bitbucket"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	slog.Info("Reading projects")

	connector := bitbucket.GetConnector(ctx)

	res, err := connector.Read(ctx, common.ReadParams{
		ObjectName: "projects",
		Fields:     datautils.NewStringSet("type", "owner", "is_private", "uuid"),
	})
	if err != nil {
		slog.Error(err.Error())
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading repos")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "repositories",
		Fields:     datautils.NewStringSet("full_name", "type", "name", "description"),
		Since:      time.Now().Add(-10 * time.Hour),
	})
	if err != nil {
		slog.Error(err.Error())
	}

	utils.DumpJSON(res, os.Stdout)
}
