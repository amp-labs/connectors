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

	slog.Info("Product created successfully")
	utils.DumpJSON(productResult, os.Stdout)

	productID := productResult.RecordId
	slog.Info("Product ID", "id", productID)

	// Test 2: Create product options (Color and Size)
	slog.Info("=== Test 2: Creating product options ===")

	optionsResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "productOptions",
		RecordData: map[string]any{
			"productId": productID,
			"options": []map[string]any{
				{
					"name": "Color",
					"values": []map[string]any{
						{"name": "Red"},
						{"name": "Blue"},
						{"name": "Green"},
					},
				},
				{
					"name": "Size",
					"values": []map[string]any{
						{"name": "Small"},
						{"name": "Medium"},
						{"name": "Large"},
					},
				},
				{
					"name": "Material",
					"values": []map[string]any{
						{"name": "Cotton"},
						{"name": "Polyester"},
					},
				},
			},
		},
	})
	if err != nil {
		slog.Error("Error creating product options", "error", err)
		// Continue to cleanup even if options creation fails
	} else {
		slog.Info("Product options created successfully")
		utils.DumpJSON(optionsResult, os.Stdout)
	}

	// Extract option IDs from the response for deletion test
	var optionIDs []string
	if optionsResult != nil && optionsResult.Data != nil {
		if options, ok := optionsResult.Data["options"].([]any); ok {
			for _, opt := range options {
				if optMap, ok := opt.(map[string]any); ok {
					if id, ok := optMap["id"].(string); ok {
						optionIDs = append(optionIDs, id)
						slog.Info("Option ID", "id", id)
					}
				}
			}
		}
	}

	// Test 3: Delete one option (Material - the last one)
	if len(optionIDs) > 0 {
		slog.Info("=== Test 3: Deleting one product option ===")

		// Delete the last option (Material)
		optionToDelete := optionIDs[len(optionIDs)-1]
		slog.Info("Deleting option", "optionId", optionToDelete)

		// For productOptions delete, RecordId format is "productId|optionId1,optionId2,..."
		deleteOptionId := productID + "|" + optionToDelete
		deleteOptionResult, err := conn.Delete(ctx, common.DeleteParams{
			ObjectName: "productOptions",
			RecordId:   deleteOptionId,
		})
		if err != nil {
			slog.Error("Error deleting product option", "error", err)
			// Continue to cleanup
		} else {
			slog.Info("Product option deleted successfully")
			utils.DumpJSON(deleteOptionResult, os.Stdout)
		}
	} else {
		slog.Info("=== Test 3: Skipping option deletion (no options found) ===")
	}

	// Test 4: Update the Product
	slog.Info("=== Test 4: Updating the product ===")

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
		// Continue to cleanup
	} else {
		slog.Info("Product updated successfully")
		utils.DumpJSON(updateResult, os.Stdout)
	}

	// Test 5: Delete the Product (cleanup)
	slog.Info("=== Test 5: Deleting the product ===")

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

	slog.Info("=== All product tests completed successfully ===")

	return 0
}
