package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/bambooHR"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	utils.SetupLogging()

	conn := bambooHR.GetBambooHRConnector(ctx)

	objects := []string{
		"employees_directory",
		"meta_users",
		"employees",
		"time_off_requests",
		"whos_out",
		"applicant_tracking_jobs",
		"meta_fields",
		"webhooks",
	}

	result, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		slog.Error("ListObjectMetadata failed", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(result, os.Stdout)
}
