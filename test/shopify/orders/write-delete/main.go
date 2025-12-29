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

// Note: creating an order requires a ProductVariant ID. To keep this runnable
// out of the box, we create a temporary product first and use its first variant.
func MainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := shopify.GetShopifyConnector(ctx)

	timestamp := time.Now().Unix()
	productTitle := fmt.Sprintf("Integration Test Order Product %d", timestamp)
	orderEmail := fmt.Sprintf("integration-order-%d@example.com", timestamp)

	// ------------------------------------------------------------
	// Step 0: Create a product to obtain a variant ID for orderCreate
	// ------------------------------------------------------------
	slog.Info("=== Step 0: Creating a product (used for line item variant) ===")

	productResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "products",
		RecordData: map[string]any{
			"title":           productTitle,
			"descriptionHtml": "<p>Temporary product created by order integration test. Safe to delete.</p>",
			"productType":     "Test",
			"vendor":          "Integration Test Vendor",
			"tags":            []string{"test", "integration", "auto-generated", "orders"},
		},
	})
	if err != nil {
		slog.Error("Error creating product", "error", err)
		return 1
	}

	slog.Info("Product created successfully")
	utils.DumpJSON(productResult, os.Stdout)

	productID := productResult.RecordId
	variantID, ok := extractFirstVariantID(productResult.Data)
	if !ok {
		slog.Error("Failed to extract first variant id from product create response")
		// best-effort cleanup
		_ = cleanupDelete(ctx, conn, "products", productID)
		return 1
	}

	// Ensure we try to cleanup the product even if later steps fail.
	defer func() {
		if productID != "" {
			_ = cleanupDelete(ctx, conn, "products", productID)
		}
	}()

	slog.Info("Using variant id for orderCreate", "variantId", variantID)

	// ------------------------------------------------------------
	// Step 1: Create an order
	// ------------------------------------------------------------
	slog.Info("=== Step 1: Creating an order ===")

	orderCreateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "orders",
		RecordData: map[string]any{
			// Minimal payload: a line item referencing a variant.
			"lineItems": []any{
				map[string]any{
					"quantity":  1,
					"variantId": variantID,
				},
			},
			// Optional, but helps identify test-generated orders in Shopify admin.
			"email": orderEmail,
			"note":  "Integration test order - safe to delete",
		},
	})
	if err != nil {
		slog.Error("Error creating order", "error", err)
		return 1
	}

	slog.Info("Order created successfully")
	utils.DumpJSON(orderCreateResult, os.Stdout)

	orderID := orderCreateResult.RecordId
	if orderID == "" {
		slog.Error("Order create returned empty RecordId")
		return 1
	}

	// Ensure we try to cleanup the order even if later steps fail.
	defer func() {
		if orderID != "" {
			_ = cleanupDelete(ctx, conn, "orders", orderID)
		}
	}()

	// ------------------------------------------------------------
	// Step 2: Update the order
	// ------------------------------------------------------------
	slog.Info("=== Step 2: Updating the order ===")

	orderUpdateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "orders",
		RecordId:   orderID,
		RecordData: map[string]any{
			"note": "Updated via integration test",
		},
	})
	if err != nil {
		slog.Error("Error updating order", "error", err)
		return 1
	}

	slog.Info("Order updated successfully")
	utils.DumpJSON(orderUpdateResult, os.Stdout)

	// ------------------------------------------------------------
	// Step 3: Close the order
	// ------------------------------------------------------------
	slog.Info("=== Step 3: Closing the order ===")

	orderCloseResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "orders",
		RecordId:   orderID,
		RecordData: map[string]any{
			// Shopify connector interprets status=closed as orderClose.
			"status": "closed",
		},
	})
	if err != nil {
		slog.Error("Error closing order", "error", err)
		return 1
	}

	slog.Info("Order closed successfully")
	utils.DumpJSON(orderCloseResult, os.Stdout)

	// ------------------------------------------------------------
	// Step 4: Delete the order (cleanup)
	// ------------------------------------------------------------
	slog.Info("=== Step 4: Deleting the order ===")

	orderDeleteResult, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "orders",
		RecordId:   orderID,
	})
	if err != nil {
		slog.Error("Error deleting order", "error", err)
		return 1
	}

	slog.Info("Order deleted successfully")
	utils.DumpJSON(orderDeleteResult, os.Stdout)

	// Defer cleanup would retry delete; clear it since we succeeded.
	orderID = ""

	slog.Info("=== All order tests completed successfully ===")
	return 0
}

func cleanupDelete(ctx context.Context, conn interface {
	Delete(context.Context, common.DeleteParams) (*common.DeleteResult, error)
}, objectName, recordID string) error {
	if recordID == "" {
		return nil
	}

	_, err := conn.Delete(ctx, common.DeleteParams{ObjectName: objectName, RecordId: recordID})
	if err != nil {
		slog.Warn("cleanup delete failed", "object", objectName, "recordId", recordID, "error", err)
		return err
	}

	slog.Info("cleanup delete succeeded", "object", objectName, "recordId", recordID)
	return nil
}

func extractFirstVariantID(productData map[string]any) (string, bool) {
	if productData == nil {
		return "", false
	}

	variantsRaw, ok := productData["variants"]
	if !ok {
		return "", false
	}

	variantsMap, ok := variantsRaw.(map[string]any)
	if !ok {
		return "", false
	}

	nodesRaw, ok := variantsMap["nodes"]
	if !ok {
		return "", false
	}

	nodes, ok := nodesRaw.([]any)
	if !ok || len(nodes) == 0 {
		return "", false
	}

	first, ok := nodes[0].(map[string]any)
	if !ok {
		return "", false
	}

	variantID, _ := first["id"].(string)
	if variantID == "" {
		return "", false
	}

	return variantID, true
}
