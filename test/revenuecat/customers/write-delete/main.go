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

	slog.Info("=== customers (create -> delete) ===")

	customerID, err := createCustomer(ctx, conn)
	if err != nil {
		slog.Error("Failed to create customer", "error", err)
		return 1
	}
	defer func() {
		if customerID != "" {
			if err := deleteByID(ctx, conn, "customers", customerID); err != nil {
				slog.Warn("Cleanup delete failed", "object", "customers", "id", customerID, "error", err)
			}
		}
	}()

	if err := deleteByID(ctx, conn, "customers", customerID); err != nil {
		slog.Error("Failed to delete customer", "error", err, "customer_id", customerID)
		return 1
	}
	customerID = ""

	slog.Info("RevenueCat customers write-delete test completed successfully")
	return 0
}

func createCustomer(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	customerID := fmt.Sprintf("amp-wd-%s", gofakeit.UUID())
	slog.Info("Creating customer", "id", customerID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"id": customerID,
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("customer create returned empty RecordId")
	}
	return res.RecordId, nil
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
