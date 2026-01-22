package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/test/quickbooks"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := quickbooks.GetQuickBooksConnector(ctx)

	// Test objects that support custom fields
	slog.Info("Testing metadata for objects WITH custom fields (customer, invoice)...")
	m, err := connector.ListObjectMetadata(ctx, []string{"customer", "invoice"})
	if err != nil {
		utils.Fail("error fetching metadata", "error", err)
	}

	fmt.Println("\n=== Customer Metadata ===")
	if customerMeta, ok := m.Result["customer"]; ok {
		fmt.Printf("DisplayName: %s\n", customerMeta.DisplayName)
		fmt.Printf("Total Fields: %d\n", len(customerMeta.Fields))
		fmt.Println("\nFields:")
		for fieldName, fieldMeta := range customerMeta.Fields {
			fmt.Printf("  - %s: %s (%s)\n", fieldName, fieldMeta.ValueType, fieldMeta.ProviderType)
		}
	} else if err, ok := m.Errors["customer"]; ok {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("\n=== Invoice Metadata ===")
	if invoiceMeta, ok := m.Result["invoice"]; ok {
		fmt.Printf("DisplayName: %s\n", invoiceMeta.DisplayName)
		fmt.Printf("Total Fields: %d\n", len(invoiceMeta.Fields))
		fmt.Println("\nFields:")
		for fieldName, fieldMeta := range invoiceMeta.Fields {
			fmt.Printf("  - %s: %s (%s)\n", fieldName, fieldMeta.ValueType, fieldMeta.ProviderType)
		}
	} else if err, ok := m.Errors["invoice"]; ok {
		fmt.Printf("Error: %v\n", err)
	}

	// Test objects that DON'T support custom fields
	slog.Info("\nTesting metadata for objects WITHOUT custom fields (account, item)...")
	m2, err := connector.ListObjectMetadata(ctx, []string{"account", "item"})
	if err != nil {
		utils.Fail("error fetching metadata", "error", err)
	}

	fmt.Println("\n=== Account Metadata ===")
	if accountMeta, ok := m2.Result["account"]; ok {
		fmt.Printf("DisplayName: %s\n", accountMeta.DisplayName)
		fmt.Printf("Total Fields: %d\n", len(accountMeta.Fields))
		fmt.Println("(Should only have built-in fields, no custom fields)")
	} else if err, ok := m2.Errors["account"]; ok {
		fmt.Printf("Error: %v\n", err)
	}

	// Print full JSON for inspection
	slog.Info("\n=== Full JSON Output ===")
	utils.DumpJSON(m, os.Stdout)
}
