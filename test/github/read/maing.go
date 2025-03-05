package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/github"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	if err := run(); err != nil {
		utils.Fail(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	connector := github.GetGithubConnector(ctx)

	slog.Info("Reading emails")

	res, err := connector.Read(ctx, common.ReadParams{
		ObjectName: "emails",
		Fields:     datautils.NewStringSet("email", "primary"),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	slog.Info("Reading repos")

	res, err = connector.Read(ctx, common.ReadParams{
		ObjectName: "repos",
		Fields:     datautils.NewStringSet("full_name", "downloads_url", "description", "default_branch"),
		Since:      time.Date(2024, 03, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
