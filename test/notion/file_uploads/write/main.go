package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/notion"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetNotionConnector(ctx)

	slog.Info("=== Creating file upload ===")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "file_uploads",
		RecordData: map[string]any{
			"mode":         "single_part",
			"filename":     "ampersand-test.txt",
			"content_type": "text/plain",
		},
	})
	if err != nil {
		utils.Fail("error creating Notion file upload", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
}
