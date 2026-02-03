package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/phoneburner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPhoneBurnerConnector(ctx)

	slog.Info("=== Reading contacts ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "contacts",
		Fields:     connectors.Fields("contact_user_id", "first_name", "last_name", "raw_phone"),
	})

	slog.Info("=== Reading members ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "members",
		Fields:     connectors.Fields("user_id", "username", "email_address"),
	})

	slog.Info("=== Reading voicemails ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "voicemails",
		Fields:     connectors.Fields("recording_id", "name", "created_when"),
	})

	// Folders is not paginated and returns a map shape, but the connector normalizes it to rows.
	slog.Info("=== Reading folders ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "folders",
		Fields:     connectors.Fields("folder_id", "folder_name"),
	})

	slog.Info("=== Reading customfields ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "customfields",
		Fields:     connectors.Fields("custom_field_id", "display_name", "type_name"),
	})

	slog.Info("=== Reading dialsession ===")
	testscenario.ReadThroughPages(ctx, conn, common.ReadParams{
		ObjectName: "dialsession",
		Fields:     connectors.Fields("dialsession_id", "start_when", "call_count"),
	})
}
