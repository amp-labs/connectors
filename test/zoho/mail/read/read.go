package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/zoho"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := zoho.GetZohoConnector(ctx, providers.ModuleZohoMail)
	if _, err := conn.GetPostAuthInfo(ctx); err != nil {
		utils.Fail(err.Error())
	}

	slog.Info("Reading notes..")
	notes, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "notes",
		Fields:     connectors.Fields("entityId", "title", "content"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading from Zoho Mail", "error", err)
	}

	utils.DumpJSON(notes, os.Stdout)

	if notes.NextPage != "" {
		slog.Info("Reading next page of notes..", "nextPage", notes.NextPage)

		nextNotes, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "notes",
			Fields:     connectors.Fields("entityId", "title", "content"),
			NextPage:   notes.NextPage,
		})
		if err != nil {
			utils.Fail("error reading next page from Zoho Mail", "error", err)
		}

		utils.DumpJSON(nextNotes, os.Stdout)
	}

	slog.Info("Reading tasks..")
	tasks, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "tasks",
		Fields:     connectors.Fields("id", "title", "status"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading from Zoho Mail", "error", err)
	}

	utils.DumpJSON(tasks, os.Stdout)

	slog.Info("Reading messages..")
	messages, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "messages",
		Fields:     connectors.Fields("messageId", "subject", "fromAddress"),
		PageSize:   2,
	})
	if err != nil {
		utils.Fail("error reading from Zoho Mail", "error", err)
	}

	utils.DumpJSON(messages, os.Stdout)

	if messages.NextPage != "" {
		slog.Info("Reading next page of messages..", "nextPage", messages.NextPage)

		nextMessages, err := conn.Read(ctx, common.ReadParams{
			ObjectName: "messages",
			Fields:     connectors.Fields("messageId", "subject", "fromAddress"),
			NextPage:   messages.NextPage,
		})
		if err != nil {
			utils.Fail("error reading next page from Zoho Mail", "error", err)
		}

		utils.DumpJSON(nextMessages, os.Stdout)
	}

	slog.Info("Read operation completed successfully.")
}
