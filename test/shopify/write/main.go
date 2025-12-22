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
	// Generate unique phone using last 7 digits of timestamp
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

	slog.Info("Customer created successfully")
	utils.DumpJSON(customerResult, os.Stdout)

	customerID := customerResult.RecordId
	slog.Info("Customer ID", "id", customerID)

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

	// Test 3: Create an Address for the Customer
	slog.Info("=== Test 3: Creating an address for the customer ===")

	addressResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customerAddresses",
		RecordData: map[string]any{
			"customerId": customerID,
			"address": map[string]any{
				"address1":     "123 Test Street",
				"address2":     "Suite 100",
				"city":         "Toronto",
				"company":      "Test Company Inc",
				"countryCode":  "CA",
				"firstName":    "Test",
				"lastName":     "Customer",
				"phone":        "+14165551234",
				"provinceCode": "ON",
				"zip":          "M5V 1J1",
			},
			"setAsDefault": true,
		},
	})
	if err != nil {
		slog.Error("Error creating address", "error", err)
		return 1
	}

	slog.Info("Address created successfully")
	utils.DumpJSON(addressResult, os.Stdout)

	addressID := addressResult.RecordId
	slog.Info("Address ID", "id", addressID)

	// Test 4: Update the Address
	slog.Info("=== Test 4: Updating the address ===")

	addressUpdateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customerAddresses",
		RecordId:   addressID,
		RecordData: map[string]any{
			"customerId": customerID,
			"address": map[string]any{
				"address1":     "456 Updated Avenue",
				"address2":     "Floor 5",
				"city":         "Vancouver",
				"company":      "Updated Company Inc",
				"countryCode":  "CA",
				"firstName":    "Updated",
				"lastName":     "Customer",
				"phone":        "+16045551234",
				"provinceCode": "BC",
				"zip":          "V6B 1A1",
			},
		},
	})
	if err != nil {
		slog.Error("Error updating address", "error", err)
		return 1
	}

	slog.Info("Address updated successfully")
	utils.DumpJSON(addressUpdateResult, os.Stdout)

	// Test 5: Create a Second Address
	slog.Info("=== Test 5: Creating a second address ===")

	secondAddressResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customerAddresses",
		RecordData: map[string]any{
			"customerId": customerID,
			"address": map[string]any{
				"address1":     "789 Secondary Street",
				"city":         "Ottawa",
				"countryCode":  "CA",
				"firstName":    "Test",
				"lastName":     "Customer",
				"provinceCode": "ON",
				"zip":          "K1A 0A1",
			},
			"setAsDefault": false,
		},
	})
	if err != nil {
		slog.Error("Error creating second address", "error", err)
		return 1
	}

	slog.Info("Second address created successfully")
	utils.DumpJSON(secondAddressResult, os.Stdout)

	secondAddressID := secondAddressResult.RecordId
	slog.Info("Second Address ID", "id", secondAddressID)

	// Test 6: Update Default Address
	slog.Info("=== Test 6: Setting the second address as default ===")

	defaultAddressResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customerDefaultAddress",
		RecordData: map[string]any{
			"customerId": customerID,
			"addressId":  secondAddressID,
		},
	})
	if err != nil {
		slog.Error("Error updating default address", "error", err)
		return 1
	}

	slog.Info("Default address updated successfully")
	utils.DumpJSON(defaultAddressResult, os.Stdout)

	// Test 7: Delete the first address (not the default anymore after Test 6)
	slog.Info("=== Test 7: Deleting the first address (non-default) ===")

	// For customerAddresses delete, RecordId format is "customerId|addressId"
	firstAddressDeleteId := customerID + "|" + addressID
	deleteFirstAddressResult, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "customerAddresses",
		RecordId:   firstAddressDeleteId,
	})
	if err != nil {
		slog.Error("Error deleting first address", "error", err)
		return 1
	}

	slog.Info("First address deleted successfully")
	utils.DumpJSON(deleteFirstAddressResult, os.Stdout)

	// Note: Second address is now the default and only address, so we skip deleting it.
	// Deleting the customer will clean up remaining addresses.

	// Test 8: Delete the customer
	slog.Info("=== Test 8: Deleting the customer ===")

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

	slog.Info("=== All write tests completed successfully ===")

	return 0
}
