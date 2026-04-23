package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	slog.Info("=== Reading all contacts ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contact_user_id", "first_name", "last_name", "raw_phone"),
	})

	// Incremental read — verifies PST timezone handling: Since=now-1h is always
	// 7h ahead of PST "now", so this fails without the update_to fix in read.go.
	slog.Info("=== Reading contacts incrementally (Since=now-1h) ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contact_user_id", "first_name", "last_name"),
		Since:      time.Now().UTC().Add(-1 * time.Hour),
	})
}
