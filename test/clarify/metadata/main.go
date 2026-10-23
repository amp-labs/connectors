package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/providers/clarify"
	connTest "github.com/amp-labs/connectors/test/clarify"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	utils.SetupLogging()

	conn := connTest.GetClarifyConnector(ctx)

	result, err := conn.ListObjectMetadata(ctx, clarify.SupportedReadObjects())
	if err != nil {
		slog.Error("ListObjectMetadata failed", "error", err)
		os.Exit(1)
	}

	utils.DumpJSON(result, os.Stdout)
}
