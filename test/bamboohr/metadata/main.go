package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/bamboohr"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	utils.SetupLogging()

	conn := bamboohr.GetBambooHRConnector(ctx)

	objects := []string{
		"employees/directory",
		"meta/users",
		"employees",
		"time-off/requests",
		"whos-out",
		"applicant_tracking/jobs",
		"meta/fields",
		"webhooks",
	}

	result, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		slog.Error("ListObjectMetadata failed", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(result, os.Stdout)
}
