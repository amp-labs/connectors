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

	// Generate unique email and phone to avoid conflicts
	timestamp := time.Now().Unix()
	testEmail := fmt.Sprintf("testcustomer%d@example.com", timestamp)
	testPhone := fmt.Sprintf("+1646555%04d", timestamp%10000)

	// Test 1: Create a Customer
	slog.Info("=== Test 1: Creating a customer ===")

	customerResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"email":     testEmail,
			"firstName": "Test",
			"lastName":  "Customer",
			"phone":     testPhone,
			"note":      "Integration test customer - safe to delete",
		},
	})
	if err != nil {
		slog.Error("Error creating customer", "error", err)
		return 1
	}

	customerID := customerResult.RecordId
	slog.Info("Customer created successfully", "id", customerID)
	utils.DumpJSON(customerResult, os.Stdout)

	// Test 2: Update the Customer
	slog.Info("=== Test 2: Updating the customer ===")

	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordId:   customerID,
		RecordData: map[string]any{
			"firstName": "Updated",
			"lastName":  "TestCustomer",
			"note":      "Updated via integration test",
		},
	})
	if err != nil {
		slog.Error("Error updating customer", "error", err)
		return 1
	}

	slog.Info("Customer updated successfully")
	utils.DumpJSON(updateResult, os.Stdout)

	// Test 3: Delete the customer
	slog.Info("=== Test 3: Deleting the customer ===")

	deleteCustomerResult, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "customers",
		RecordId:   customerID,
	})
	if err != nil {
		slog.Error("Error deleting customer", "error", err)
		return 1
	}

	slog.Info("Customer deleted successfully")
	utils.DumpJSON(deleteCustomerResult, os.Stdout)

	return 0
}
