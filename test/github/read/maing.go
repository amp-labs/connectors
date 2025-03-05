package main

import (
	"context"
	"log/slog"
	"os"

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
	})
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}
