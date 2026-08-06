// Live write/delete test for the Mailgun connector.
//
// Exercises every write style against the real API and cleans up after itself:
//   - lists: create -> update -> delete (form-encoded, account-scoped)
//   - lists/members: create + upsert-update under a parent list (sidecar
//     list_address), cleaned up by deleting the parent list
//   - routes: create -> update -> delete (form-encoded)
//   - forwards: create -> delete (query-parameter write)
//   - templates: create -> update -> delete (form-encoded, domain-scoped)
//   - bounces: create -> delete (suppression: no update endpoint)
//   - messages: send in test mode (o:testmode) to the account owner
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/mailgun"
	testMailgun "github.com/amp-labs/connectors/test/mailgun"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testMailgun.GetMailgunConnector(ctx)

	testListsAndMembers(ctx, conn)
	testRoutes(ctx, conn)
	testForwards(ctx, conn)
	testTemplates(ctx, conn)
	testBounces(ctx, conn)
	testWebhooks(ctx, conn)
	testMessages(ctx, conn)
}

func write(ctx context.Context, conn *mailgun.Connector, params common.WriteParams) *common.WriteResult {
	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error("write failed", "object", params.ObjectName, "error", err.Error())

		return nil
	}

	slog.Info("write ok", "object", params.ObjectName, "recordId", res.RecordId)
	utils.DumpJSON(res, os.Stdout)

	return res
}

func remove(ctx context.Context, conn *mailgun.Connector, objectName, recordID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{ObjectName: objectName, RecordId: recordID})
	if err != nil {
		slog.Error("delete failed", "object", objectName, "recordId", recordID, "error", err.Error())

		return
	}

	slog.Info("delete ok", "object", objectName, "recordId", recordID, "success", res.Success)
}

func testListsAndMembers(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== lists: create -> update -> members -> delete ===")

	domain := listDomain(conn)
	listAddress := "amp-live-test@" + domain

	created := write(ctx, conn, common.WriteParams{
		ObjectName: "lists",
		RecordData: map[string]any{
			"address": listAddress,
			"name":    "Ampersand live test list",
		},
	})
	if created == nil {
		return
	}

	defer remove(ctx, conn, "lists", listAddress)

	write(ctx, conn, common.WriteParams{
		ObjectName: "lists",
		RecordId:   listAddress,
		RecordData: map[string]any{"description": "updated by live test"},
	})

	// Member create, then upsert-update through the same POST endpoint.
	write(ctx, conn, common.WriteParams{
		ObjectName: "lists/members",
		RecordData: map[string]any{
			"list_address": listAddress,
			"address":      "amp-member@example.com",
			"name":         "Live Test Member",
		},
	})

	write(ctx, conn, common.WriteParams{
		ObjectName: "lists/members",
		RecordId:   "amp-member@example.com",
		RecordData: map[string]any{
			"list_address": listAddress,
			"name":         "Live Test Member (updated)",
		},
	})
}

func testRoutes(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== routes: create -> update -> delete ===")

	created := write(ctx, conn, common.WriteParams{
		ObjectName: "routes",
		RecordData: map[string]any{
			"expression":  "match_recipient('amp-live-test@example.com')",
			"action":      "stop()",
			"description": "ampersand live test route",
			"priority":    5,
		},
	})
	if created == nil {
		return
	}

	defer remove(ctx, conn, "routes", created.RecordId)

	write(ctx, conn, common.WriteParams{
		ObjectName: "routes",
		RecordId:   created.RecordId,
		RecordData: map[string]any{"description": "ampersand live test route (updated)"},
	})
}

func testForwards(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== forwards: create -> delete (query-parameter write) ===")

	created := write(ctx, conn, common.WriteParams{
		ObjectName: "forwards",
		RecordData: map[string]any{
			"match":       "amp-live-forward@" + listDomain(conn),
			"forward.url": "https://example.com/amp-live-test",
		},
	})
	if created == nil {
		return
	}

	remove(ctx, conn, "forwards", created.RecordId)
}

func testTemplates(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== templates: create -> update -> delete (domain-scoped) ===")

	created := write(ctx, conn, common.WriteParams{
		ObjectName: "templates",
		RecordData: map[string]any{
			"name":        "amp-live-test-template",
			"description": "ampersand live test",
			"template":    "<html><body>Hello {{name}}</body></html>",
		},
	})
	if created == nil {
		return
	}

	defer remove(ctx, conn, "templates", "amp-live-test-template")

	write(ctx, conn, common.WriteParams{
		ObjectName: "templates",
		RecordId:   "amp-live-test-template",
		RecordData: map[string]any{"description": "ampersand live test (updated)"},
	})
}

func testBounces(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== bounces: create -> delete (suppression, no update) ===")

	created := write(ctx, conn, common.WriteParams{
		ObjectName: "bounces",
		RecordData: map[string]any{
			"address": "amp-live-bounce@example.com",
			"code":    "550",
			"error":   "ampersand live test",
		},
	})
	if created == nil {
		return
	}

	remove(ctx, conn, "bounces", "amp-live-bounce@example.com")
}

func testWebhooks(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== webhooks: create -> update -> delete (flat webhook_id response) ===")

	created := write(ctx, conn, common.WriteParams{
		ObjectName: "webhooks",
		RecordData: map[string]any{
			"url":         "https://example.com/amp-live-webhook",
			"event_types": "delivered",
			"description": "ampersand live test webhook",
		},
	})
	if created == nil {
		return
	}

	defer remove(ctx, conn, "webhooks", created.RecordId)

	write(ctx, conn, common.WriteParams{
		ObjectName: "webhooks",
		RecordId:   created.RecordId,
		RecordData: map[string]any{
			"url":         "https://example.com/amp-live-webhook-updated",
			"event_types": "delivered",
		},
	})
}

func testMessages(ctx context.Context, conn *mailgun.Connector) {
	slog.Info("=== messages: send in test mode ===")

	domain := listDomain(conn)

	// o:testmode accepts the message without delivering it. Sandbox domains can
	// only address authorized recipients; the account owner qualifies.
	write(ctx, conn, common.WriteParams{
		ObjectName: "messages",
		RecordData: map[string]any{
			"from":       fmt.Sprintf("Ampersand Live Test <postmaster@%v>", domain),
			"to":         "integration.user+mailgun@withampersand.com",
			"subject":    "Ampersand connector live test",
			"text":       "Sent by the mailgun connector write live test.",
			"o:testmode": "yes",
		},
	})
}

// listDomain returns the workspace domain from the credentials file.
func listDomain(_ *mailgun.Connector) string {
	return testMailgun.GetWorkspace()
}
