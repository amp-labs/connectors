package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/revenuecat"
	connTest "github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRevenueCatConnector(ctx)

	slog.Info("=== integrations_webhooks (create -> update -> delete) ===")

	webhookID, err := createWebhook(ctx, conn)
	if err != nil {
		slog.Error("Failed to create webhook", "error", err)
		return 1
	}
	defer func() {
		if webhookID != "" {
			if err := deleteByID(ctx, conn, "integrations_webhooks", webhookID); err != nil {
				slog.Warn("Cleanup delete failed", "object", "integrations_webhooks", "id", webhookID, "error", err)
			}
		}
	}()

	if err := updateWebhook(ctx, conn, webhookID); err != nil {
		slog.Error("Failed to update webhook", "error", err, "webhook_id", webhookID)
		return 1
	}

	if err := deleteByID(ctx, conn, "integrations_webhooks", webhookID); err != nil {
		slog.Error("Failed to delete webhook", "error", err, "webhook_id", webhookID)
		return 1
	}
	webhookID = ""

	slog.Info("RevenueCat integrations_webhooks write-delete test completed successfully")
	return 0
}

func createWebhook(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	name := fmt.Sprintf("Amp WD Webhook %s", gofakeit.Word())
	slog.Info("Creating webhook", "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "integrations_webhooks",
		RecordData: map[string]any{
			"name": name,
			"url":  fmt.Sprintf("https://example.com/amp-wd-%s", gofakeit.UUID()),
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("webhook create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateWebhook(ctx context.Context, conn *revenuecat.Connector, webhookID string) error {
	slog.Info("Updating webhook", "webhook_id", webhookID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "integrations_webhooks",
		RecordId:   webhookID,
		RecordData: map[string]any{
			"name": fmt.Sprintf("Amp WD Webhook Updated %s", gofakeit.Word()),
		},
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	return nil
}

func deleteByID(ctx context.Context, conn *revenuecat.Connector, objectName, recordID string) error {
	slog.Info("Deleting record", "object", objectName, "id", recordID)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   recordID,
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)

	if !res.Success {
		return fmt.Errorf("delete reported Success=false")
	}
	return nil
}
