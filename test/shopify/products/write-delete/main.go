package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/shopify"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := shopify.GetShopifyConnector(ctx)

	// Generate unique product title to avoid conflicts
	timestamp := time.Now().Unix()
	productTitle := fmt.Sprintf("Test Product %d", timestamp)

	// Test 1: Create a Product
	slog.Info("=== Test 1: Creating a product ===")

	productResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"title":           productTitle,
			"descriptionHtml": "<p>This is a test product created by integration tests. Safe to delete.</p>",
			"productType":     "Test",
			"vendor":          "Integration Test Vendor",
			"tags":            []string{"test", "integration", "auto-generated"},
		},
	})
	if err != nil {
		slog.Error("Error creating product", "error", err)
		return 1
	}

	productID := productResult.RecordId
	slog.Info("Product created successfully", "id", productID)
	utils.DumpJSON(productResult, os.Stdout)

	// Test 2: Update the Product
	slog.Info("=== Test 2: Updating the product ===")

	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordId:   productID,
		RecordData: map[string]any{
			"title":           productTitle + " (Updated)",
			"descriptionHtml": "<p>Updated description. This product was modified by integration tests.</p>",
			"status":          "ACTIVE",
		},
	})
	if err != nil {
		slog.Error("Error updating product", "error", err)
		return 1
	}

	slog.Info("Product updated successfully")
	utils.DumpJSON(updateResult, os.Stdout)

	// Test 3: Delete the Product (cleanup)
	slog.Info("=== Test 3: Deleting the product ===")

	deleteResult, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "products",
		RecordId:   productID,
	})
	if err != nil {
		slog.Error("Error deleting product", "error", err)
		return 1
	}

	slog.Info("Product deleted successfully")
	utils.DumpJSON(deleteResult, os.Stdout)

	return 0
}
