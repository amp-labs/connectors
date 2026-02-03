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

	// WARNING: phone number delete is very destructive (deletes the number).
	// We intentionally do not run it by default.
	if os.Getenv("PHONEBURNER_ALLOW_PHONE_NUMBER_DELETE") != "true" {
		slog.Info("Skipping phonenumber delete (set PHONEBURNER_ALLOW_PHONE_NUMBER_DELETE=true to run)")
		return
	}

	phoneToDelete := os.Getenv("PHONEBURNER_PHONE_NUMBER_TO_DELETE")
	if phoneToDelete == "" {
		utils.Fail("PHONEBURNER_PHONE_NUMBER_TO_DELETE must be set when PHONEBURNER_ALLOW_PHONE_NUMBER_DELETE=true")
	}

	slog.Warn("Deleting phone number (destructive)", "phone_number", phoneToDelete)
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "phonenumber",
		RecordId:   phoneToDelete,
	})
	if err != nil {
		utils.Fail("Failed to delete phone number", "error", err, "phone_number", phoneToDelete)
	}
	utils.DumpJSON(res, os.Stdout)

	slog.Info("PhoneBurner phonenumber delete test completed successfully")
}

