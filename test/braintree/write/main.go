package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/braintree"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	conn := braintree.GetBraintreeConnector(ctx)

	fmt.Println("=== Testing Braintree Write Operations ===")
	fmt.Println()

	// 1. Create a Customer
	fmt.Println("1. Creating a customer...")

	customerResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordData: map[string]any{
			"firstName": "Test",
			"lastName":  "User",
			"email":     "testuser@example.com",
			"company":   "Test Company",
		},
	})
	if err != nil {
		log.Fatal("Error creating customer:", err)
	}

	fmt.Println("Customer created successfully!")
	utils.DumpJSON(customerResult, os.Stdout)
	fmt.Println()

	customerID := customerResult.RecordId
	fmt.Printf("Customer ID: %s\n\n", customerID)

	// 2. Update the Customer
	fmt.Println("2. Updating the customer...")

	updateResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordId:   customerID,
		RecordData: map[string]any{
			"firstName": "Updated",
			"lastName":  "Customer",
			"email":     "updated@example.com",
		},
	})
	if err != nil {
		log.Fatal("Error updating customer:", err)
	}

	fmt.Println("Customer updated successfully!")
	utils.DumpJSON(updateResult, os.Stdout)
	fmt.Println()

	// 3. Create a Payment Method (vault to customer)
	fmt.Println("3. Creating a payment method (vaulting to customer)...")

	pmResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "paymentMethods",
		RecordData: map[string]any{
			"paymentMethodId": "fake-valid-nonce",
			"customerId":      customerID,
		},
	})
	if err != nil {
		fmt.Printf("Error creating payment method: %v\n", err)
	} else {
		fmt.Println("Payment method created successfully!")
		utils.DumpJSON(pmResult, os.Stdout)
	}

	fmt.Println()

	// 4. Update Payment Method (billing address)
	if pmResult != nil && pmResult.RecordId != "" {
		fmt.Println("4. Updating payment method billing address...")

		pmUpdateResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "paymentMethods",
			RecordData: map[string]any{
				"paymentMethodId": pmResult.RecordId,
				"billingAddress": map[string]any{
					"streetAddress": "123 Main St",
					"locality":      "San Francisco",
					"region":        "CA",
					"postalCode":    "94105",
				},
			},
		})
		if err != nil {
			fmt.Printf("Error updating payment method: %v\n", err)
		} else {
			fmt.Println("Payment method updated successfully!")
			utils.DumpJSON(pmUpdateResult, os.Stdout)
		}

		fmt.Println()
	}

	// 5. Create a Transaction
	fmt.Println("5. Creating a transaction (charging a payment method)...")

	txnResult, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "transactions",
		RecordData: map[string]any{
			"paymentMethodId": "fake-valid-nonce",
			"transaction": map[string]any{
				"amount": "10.00",
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating transaction: %v\n", err)
	} else {
		fmt.Println("Transaction created successfully!")
		utils.DumpJSON(txnResult, os.Stdout)
	}

	fmt.Println()

	fmt.Println("=== All Write Tests Completed Successfully ===")
}
