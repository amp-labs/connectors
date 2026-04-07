package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	// WARNING: members delete is vendor-level and destructive. Do not run by default.
	if os.Getenv("PHONEBURNER_ALLOW_MEMBER_DELETE") != "true" {
		slog.Info("Skipping members delete (set PHONEBURNER_ALLOW_MEMBER_DELETE=true to run)")
		return
	}

	memberID := os.Getenv("PHONEBURNER_MEMBER_ID_TO_DELETE")
	if memberID == "" {
		utils.Fail("PHONEBURNER_MEMBER_ID_TO_DELETE must be set when PHONEBURNER_ALLOW_MEMBER_DELETE=true")
	}

	slog.Warn("Deleting member (destructive)", "user_id", memberID)
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "members",
		RecordId:   memberID,
	})
	if err != nil {
		utils.Fail("Failed to delete member", "error", err, "user_id", memberID)
	}
	utils.DumpJSON(res, os.Stdout)

	slog.Info("PhoneBurner members delete test completed successfully")
}
