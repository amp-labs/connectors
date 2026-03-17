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

	slog.Info("=== offerings (create -> update -> delete) ===")

	offeringID, err := createOffering(ctx, conn)
	if err != nil {
		slog.Error("Failed to create offering", "error", err)
		return 1
	}
	defer func() {
		if offeringID != "" {
			if err := deleteByID(ctx, conn, "offerings", offeringID); err != nil {
				slog.Warn("Cleanup delete failed", "object", "offerings", "id", offeringID, "error", err)
			}
		}
	}()

	if err := updateOffering(ctx, conn, offeringID); err != nil {
		slog.Error("Failed to update offering", "error", err, "offering_id", offeringID)
		return 1
	}

	if err := deleteByID(ctx, conn, "offerings", offeringID); err != nil {
		slog.Error("Failed to delete offering", "error", err, "offering_id", offeringID)
		return 1
	}
	offeringID = ""

	slog.Info("RevenueCat offerings write-delete test completed successfully")
	return 0
}

func createOffering(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	lookupKey := fmt.Sprintf("amp-wd-%s", gofakeit.UUID())
	slog.Info("Creating offering", "lookup_key", lookupKey)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "offerings",
		RecordData: map[string]any{
			"lookup_key":   lookupKey,
			"display_name": "Amp WD Offering",
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("offering create returned empty RecordId")
	}
	return res.RecordId, nil
}

func updateOffering(ctx context.Context, conn *revenuecat.Connector, offeringID string) error {
	slog.Info("Updating offering", "offering_id", offeringID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "offerings",
		RecordId:   offeringID,
		RecordData: map[string]any{
			"display_name": fmt.Sprintf("Amp WD Offering Updated %s", gofakeit.Word()),
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
