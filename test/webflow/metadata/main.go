package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/webflow"
)

func main() {
	ctx := context.Background()
	utils.SetupLogging()

	conn := webflow.GetWebflowConnector(ctx)

	objects := []string{
		"sites",
		"collections",
		"pages",
		"forms",
		"form_submissions",
		"assets",
		"orders",
		"products",
		"webhooks",
		"activity_logs",
	}

	result, err := conn.ListObjectMetadata(ctx, objects)
	if err != nil {
		slog.Error("ListObjectMetadata failed", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(result, os.Stdout)
}
