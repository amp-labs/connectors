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

	slog.Info("=== entitlements (create -> update -> delete) ===")

	entitlementID, err := createEntitlement(ctx, conn)
	if err != nil {
		slog.Error("Failed to create entitlement", "error", err)
		return 1
	}
	defer func() {
		if entitlementID != "" {
			if err := deleteByID(ctx, conn, "entitlements", entitlementID); err != nil {
				slog.Warn("Cleanup delete failed", "object", "entitlements", "id", entitlementID, "error", err)
			}
		}
	}()

	if err := updateEntitlement(ctx, conn, entitlementID); err != nil {
		slog.Error("Failed to update entitlement", "error", err, "entitlement_id", entitlementID)
		return 1
	}

	if err := deleteByID(ctx, conn, "entitlements", entitlementID); err != nil {
		slog.Error("Failed to delete entitlement", "error", err, "entitlement_id", entitlementID)
		return 1
	}
	entitlementID = ""

	slog.Info("RevenueCat entitlements write-delete test completed successfully")
	return 0
}

func createEntitlement(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	lookupKey := fmt.Sprintf("amp-wd-%s", gofakeit.UUID())
	slog.Info("Creating entitlement", "lookup_key", lookupKey)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "entitlements",
		RecordData: map[string]any{
			"lookup_key":   lookupKey,
			"display_name": "Amp WD Entitlement",
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("entitlement create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateEntitlement(ctx context.Context, conn *revenuecat.Connector, entitlementID string) error {
	slog.Info("Updating entitlement", "entitlement_id", entitlementID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "entitlements",
		RecordId:   entitlementID,
		RecordData: map[string]any{
			"display_name": "Amp WD Entitlement Updated",
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
