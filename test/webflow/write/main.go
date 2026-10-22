package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/webflow"
)

// Exercises create and update against a real site. Delete is not part of this
// step, so the created webhook and collection are left behind (remove them
// with DELETE /v2/webhooks/{id} and DELETE /v2/collections/{id}).
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetWebflowConnector(ctx)

	slog.Info("> TEST Create webhook")

	webhook, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "webhooks",
		RecordData: map[string]any{
			"triggerType": "form_submission",
			"url":         "https://example.com/hooks/webflow-connector-test",
		},
	})
	if err != nil {
		utils.Fail("error creating webhook", "error", err)
	}

	utils.DumpJSON(webhook, os.Stdout)

	if webhook.RecordId == "" {
		utils.Fail("expected a record id after creating webhook")
	}

	slog.Info("> TEST Create collection")

	collection, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "collections",
		RecordData: map[string]any{
			"displayName":  "Connector Test Posts",
			"singularName": "Connector Test Post",
			"slug":         "connector-test-posts",
		},
	})
	if err != nil {
		utils.Fail("error creating collection", "error", err)
	}

	utils.DumpJSON(collection, os.Stdout)

	slog.Info("> TEST Update collection", "id", collection.RecordId)

	updated, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "collections",
		RecordId:   collection.RecordId,
		RecordData: map[string]any{"displayName": "Connector Test Posts (updated)"},
	})
	if err != nil {
		utils.Fail("error updating collection", "error", err)
	}

	utils.DumpJSON(updated, os.Stdout)

	slog.Info("> TEST Update page title and revert")

	pages, err := conn.Read(ctx, common.ReadParams{ObjectName: "pages", Fields: connectors.Fields("id", "title")})
	if err != nil || len(pages.Data) == 0 {
		utils.Fail("error reading pages", "error", err)
	}

	page := pages.Data[0]
	originalTitle, _ := page.Fields["title"].(string)

	renamed, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "pages",
		RecordId:   page.Id,
		RecordData: map[string]any{"title": originalTitle + " (connector test)"},
	})
	if err != nil {
		utils.Fail("error updating page", "error", err)
	}

	utils.DumpJSON(renamed, os.Stdout)

	if _, err = conn.Write(ctx, common.WriteParams{
		ObjectName: "pages",
		RecordId:   page.Id,
		RecordData: map[string]any{"title": originalTitle},
	}); err != nil {
		utils.Fail("error reverting page title", "error", err)
	}

	slog.Info("page title reverted", "title", originalTitle)
}
