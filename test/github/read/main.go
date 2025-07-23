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
	"github.com/amp-labs/connectors/test/github"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	slog.Info("Reading emails")

	connector := github.GetGithubConnector(ctx)

	res, err := connector.Read(ctx, common.ReadParams{
		ObjectName: "emails",
		Fields:     datautils.NewStringSet("email", "primary"),
	})

	if err != nil {
		slog.Error(err.Error())
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading repos")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "repos",
		Fields:     datautils.NewStringSet("full_name", "downloads_url", "description", "default_branch"),
		Since:      time.Date(2024, 03, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		slog.Error(err.Error())
	}

	utils.DumpJSON(res, os.Stdout)
}
