package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
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

	// Products do not support PATCH; this test creates then deletes.
	slog.Info("=== products (create -> delete) ===")

	appID, err := getFirstAppID(ctx, conn)
	if err != nil {
		slog.Error("Failed to read apps (needed for app_id)", "error", err)
		return 1
	}
	slog.Info("Using app", "id", appID)

	productID, err := createProduct(ctx, conn, appID)
	if err != nil {
		slog.Error("Failed to create product", "error", err)
		return 1
	}
	defer func() {
		if productID != "" {
			if err := deleteByID(ctx, conn, "products", productID); err != nil {
				slog.Warn("Cleanup delete failed", "object", "products", "id", productID, "error", err)
			}
		}
	}()

	if err := deleteByID(ctx, conn, "products", productID); err != nil {
		slog.Error("Failed to delete product", "error", err, "product_id", productID)
		return 1
	}
	productID = ""

	slog.Info("RevenueCat products write-delete test completed successfully")
	return 0
}

func getFirstAppID(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "apps",
		Fields:     connectors.Fields("id"),
		PageSize:   1,
	})
	if err != nil {
		return "", err
	}

	if len(res.Data) == 0 {
		return "", fmt.Errorf("no apps found in project; create an app first")
	}

	id, _ := res.Data[0].Raw["id"].(string)
	if id == "" {
		return "", fmt.Errorf("app response missing id field")
	}

	return id, nil
}

func createProduct(ctx context.Context, conn *revenuecat.Connector, appID string) (string, error) {
	storeIdentifier := fmt.Sprintf("amp.wd.%s", gofakeit.UUID())
	slog.Info("Creating product", "store_identifier", storeIdentifier, "app_id", appID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"store_identifier": storeIdentifier,
			"type":             "subscription",
			"app_id":           appID,
		},
	})
	if err != nil {
		return "", err
	}
	utils.DumpJSON(res, os.Stdout)

	if res.RecordId == "" {
		return "", fmt.Errorf("product create returned empty RecordId")
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
