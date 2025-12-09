// Package main provides an integration test for the GetRecordsByIds method.
//
// This test dynamically:
// 1. Creates 3 test accounts in Outreach
// 2. Fetches them using GetRecordsByIds with specific field selection
// 3. Verifies all created records are returned correctly
// 4. Cleans up by deleting all test accounts
//
// Run with: go run ./test/outreach/record/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/outreach"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetOutreachConnector(ctx)

	// Step 1: Create multiple test accounts
	fmt.Println("Creating test accounts...")

	recordIDs := make([]string, 0, 3)

	for i := 0; i < 3; i++ {
		writeResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "accounts",
			RecordId:   "",
			RecordData: map[string]any{
				"name":   fmt.Sprintf("Test Account %s", gofakeit.Company()),
				"domain": fmt.Sprintf("test-%s.com", gofakeit.UUID()),
			},
		})
		if err != nil {
			utils.Fail("error writing account to outreach", "error", err, "iteration", i)
		}

		recordIDs = append(recordIDs, writeResult.RecordId)
		fmt.Printf("Created account %d with ID: %s\n", i+1, writeResult.RecordId)
	}

	// Step 2: Fetch the created records using GetRecordsByIds
	fmt.Println("\nFetching accounts by IDs...")

	res, err := conn.GetRecordsByIds(ctx,
		"accounts",
		recordIDs,
		[]string{"id", "name", "domain", "industry", "numberOfEmployees"},
		nil)
	if err != nil {
		utils.Fail("error getting records by ids", "error", err)
	}

	fmt.Printf("\nSuccessfully fetched %d accounts:\n", len(res))
	utils.DumpJSON(res, os.Stdout)

	// Step 3: Verify we got all the records we created
	if len(res) != len(recordIDs) {
		utils.Fail("mismatch in record count", "expected", len(recordIDs), "got", len(res))
	}

	// Verify all record IDs are present in the results
	foundIDs := make(map[string]bool)

	for _, record := range res {
		if id, ok := record.Fields["id"].(float64); ok {
			foundIDs[fmt.Sprintf("%.0f", id)] = true
		}
	}

	for _, expectedID := range recordIDs {
		if !foundIDs[expectedID] {
			utils.Fail("expected record ID not found in results", "recordId", expectedID)
		}
	}

	fmt.Println("\n✓ All created accounts were successfully retrieved!")

	// Step 4: Clean up - delete the created accounts
	fmt.Println("\nCleaning up - deleting test accounts...")

	for i, recordID := range recordIDs {
		deleteResult, err := conn.Delete(ctx, common.DeleteParams{
			ObjectName: "accounts",
			RecordId:   recordID,
		})
		if err != nil {
			utils.Fail("error deleting account", "error", err, "recordId", recordID, "iteration", i)
		}

		if !deleteResult.Success {
			utils.Fail("delete operation failed", "recordId", recordID)
		}

		fmt.Printf("✓ Deleted account %d (ID: %s)\n", i+1, recordID)
	}

	fmt.Println("\n✓ Test completed successfully!")
	fmt.Println("✓ All test accounts have been cleaned up.")
}
