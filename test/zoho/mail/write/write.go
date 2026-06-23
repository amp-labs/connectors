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
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetZohoConnector(ctx, providers.ModuleZohoMail)
	if _, err := conn.GetPostAuthInfo(ctx); err != nil {
		utils.Fail(err.Error())
	}

	taskID := create(ctx, conn, "tasks", map[string]any{
		"title":       "Ampersand task",
		"description": "created via the Zoho Mail write connector",
	})
	update(ctx, conn, "tasks", taskID, map[string]any{
		"title": gofakeit.Sentence(3),
	})

	create(ctx, conn, "links/me", map[string]any{
		"link":  "https://withampersand.com",
		"title": "Ampersand",
	})

	create(ctx, conn, "accounts/folders", map[string]any{
		"folderName": gofakeit.Word(),
	})

	labelID := create(ctx, conn, "accounts/labels", map[string]any{
		"displayName": gofakeit.Word(),
		"color":       "#FFFFFF",
	})
	update(ctx, conn, "accounts/labels", labelID, map[string]any{
		"displayName": gofakeit.Word(),
		"color":       "#000000",
	})

	create(ctx, conn, "signature", map[string]any{
		"name":     gofakeit.Sentence(3),
		"content":  "Regards, Ampersand.",
		"position": 0,
	})

	create(ctx, conn, "customStatus", map[string]any{
		"statusName":   "Ampersand status",
		"statusColour": "#2E8BD2",
	})

	create(ctx, conn, "messages", map[string]any{
		"fromAddress": accountAddress(ctx, conn),
		"toAddress":   gofakeit.Email(),
		"subject":     gofakeit.Word(),
		"content":     gofakeit.Sentence(10),
	})

	slog.Info("Write operations completed successfully.")
}

func create(ctx context.Context, conn *zoho.Connector, objectName string, data map[string]any) string {
	slog.Info("Creating " + objectName + "..")

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordData: data,
	})
	if err != nil {
		utils.Fail("error writing to Zoho Mail", "object", objectName, "error", err)
	}

	utils.DumpJSON(result, os.Stdout)

	return result.RecordId
}

func update(ctx context.Context, conn *zoho.Connector, objectName, recordID string, data map[string]any) {
	slog.Info("Updating " + objectName + "..")

	result, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   recordID,
		RecordData: data,
	})
	if err != nil {
		utils.Fail("error updating Zoho Mail", "object", objectName, "error", err)
	}

	utils.DumpJSON(result, os.Stdout)
}

// accountAddress resolves the authenticated account's primary email address,
// which Zoho requires as the fromAddress when sending mail.
func accountAddress(ctx context.Context, conn *zoho.Connector) string {
	result, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "accounts",
		Fields:     connectors.Fields("mailboxAddress"),
	})
	if err != nil {
		utils.Fail("error reading Zoho Mail accounts", "error", err)
	}

	address, ok := result.Data[0].Fields["mailboxaddress"].(string)
	if !ok || address == "" {
		utils.Fail("account is missing a mailboxAddress")
	}

	return address
}
