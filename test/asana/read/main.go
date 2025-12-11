package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/asana"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetAsanaConnector(ctx)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "workspaces",
		Fields:     connectors.Fields("gid", "name", "email_domains"),
	})
	if err != nil {
		utils.Fail("error reading from Asana", "error", err)
	}

	slog.Info("Reading workspaces..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "users",
		Fields:     connectors.Fields("gid", "name", "email", "custom_fields"),
	})
	if err != nil {
		utils.Fail("error reading from Asana", "error", err)
	}

	slog.Info("Reading users..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "projects",
		Fields:     connectors.Fields("gid", "name", "created_at", "custom_field_settings"),
	})
	if err != nil {
		utils.Fail("error reading from Asana", "error", err)
	}

	slog.Info("Reading projects..")
	utils.DumpJSON(res, os.Stdout)
}
