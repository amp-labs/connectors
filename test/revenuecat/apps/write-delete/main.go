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

	slog.Info("=== apps (create -> update -> delete) ===")

	appID, err := createApp(ctx, conn)
	if err != nil {
		slog.Error("Failed to create app", "error", err)
		return 1
	}
	defer func() {
		if appID != "" {
			if err := deleteByID(ctx, conn, "apps", appID); err != nil {
				slog.Warn("Cleanup delete failed", "object", "apps", "id", appID, "error", err)
			}
		}
	}()

	if err := updateApp(ctx, conn, appID); err != nil {
		slog.Error("Failed to update app", "error", err, "app_id", appID)
		return 1
	}

	if err := deleteByID(ctx, conn, "apps", appID); err != nil {
		slog.Error("Failed to delete app", "error", err, "app_id", appID)
		return 1
	}
	appID = ""

	slog.Info("RevenueCat apps write-delete test completed successfully")
	return 0
}

func createApp(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	name := fmt.Sprintf("Amp WD App %s", gofakeit.Word())
	slog.Info("Creating app", "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "apps",
		RecordData: map[string]any{
			"name": name,
			"type": "app_store",
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("app create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateApp(ctx context.Context, conn *revenuecat.Connector, appID string) error {
	name := fmt.Sprintf("Amp WD App Updated %s", gofakeit.Word())
	slog.Info("Updating app", "app_id", appID, "name", name)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "apps",
		RecordId:   appID,
		RecordData: map[string]any{
			"name": name,
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
