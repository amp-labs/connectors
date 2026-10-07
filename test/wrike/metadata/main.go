package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/wrike"
)

func main() {
	ctx := context.Background()
	utils.SetupLogging()

	conn := wrike.GetWrikeConnector(ctx)

	objects := []string{
		"tasks",
		"folders",
		"contacts",
		"customfields",
		"comments",
		"timelogs",
		"workflows",
		"spaces",
		"groups",
		"approvals",
	}

	result, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		slog.Error("ListObjectMetadata failed", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(result, os.Stdout)
}
