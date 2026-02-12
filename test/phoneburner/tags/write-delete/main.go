package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	title := fmt.Sprintf("amp-tag-%s-%d", gofakeit.Word(), time.Now().Unix())

	conn := connTest.GetPhoneBurnerConnector(ctx)

	// NOTE: we don't support tags delete right now, so there is no auto-cleanup.
	slog.Warn("Creating tag (no auto-cleanup)", "title", title)
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "tags",
		RecordData: map[string]any{
			"title": title,
		},
	})
	if err != nil {
		utils.Fail("Failed to create tag", "error", err)
	}

	utils.DumpJSON(res, os.Stdout)
	if !res.Success {
		utils.Fail("Tag create returned Success=false")
	}
	if res.RecordId == "" {
		utils.Fail("Tag create returned empty RecordId")
	}

	slog.Info("PhoneBurner tags write test completed successfully", "id", res.RecordId, "title", title)
}
