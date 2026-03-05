package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesloft"
	connTest "github.com/amp-labs/connectors/test/salesloft"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesloftConnector(ctx)

	TestGetRecordsByIdsPeople(ctx, conn)

	TestGetRecordsByCalls(ctx, conn)

	fmt.Println("\nAll tests completed successfully!")
}

func TestGetRecordsByIdsPeople(ctx context.Context, conn *salesloft.Connector) {

	fmt.Println("Creating people...")

	recordIDs := make([]string, 0, 3)

	for i := 0; i < 3; i++ {
		writeResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "people",
			RecordId:   "",
			RecordData: map[string]any{
				"firstName":     fmt.Sprintf("Test Person %s", gofakeit.Name()),
				"email_address": fmt.Sprintf("test-%s@%s.com", gofakeit.UUID(), gofakeit.DomainName()),
			},
		})
		if err != nil {
			utils.Fail("error writing person to salesloft", "error", err, "iteration", i)
		}

		recordIDs = append(recordIDs, writeResult.RecordId)
		fmt.Printf("Created person %d with ID: %s\n", i+1, writeResult.RecordId)
	}

	// Step 2: Fetch the created records using GetRecordsByIds
	fmt.Println("\nFetching people by IDs...")

	res, err := conn.GetRecordsByIds(ctx, common.ReadByIdsParams{
		ObjectName: "people",
		RecordIds:  recordIDs,
		Fields:     []string{"id", "firstName", "email_address"},
	})
	if err != nil {
		utils.Fail("error getting records by ids", "error", err)
	}

	fmt.Printf("\nSuccessfully fetched %d people:\n", len(res))
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

	fmt.Println("\n✓ All created people were successfully retrieved!")

	fmt.Println("\n✓ Test completed successfully!")
	fmt.Println("✓ All test people have been cleaned up.")

}

func TestGetRecordsByCalls(ctx context.Context, conn *salesloft.Connector) {
	fmt.Println("Creating accounts...")

	recordIDs := make([]string, 0, 4)

	for i := 0; i < 3; i++ {
		writeResult, err := conn.Write(ctx, common.WriteParams{
			ObjectName: "accounts",
			RecordId:   "",
			RecordData: map[string]any{
				"name":   fmt.Sprintf("Test Account %s", gofakeit.Name()),
				"domain": gofakeit.DomainName(),
			},
		})
		if err != nil {
			utils.Fail("error writing account to salesloft", "error", err, "iteration", i)
		}

		recordIDs = append(recordIDs, writeResult.RecordId)
		fmt.Printf("Created account %d with ID: %s\n", i+1, writeResult.RecordId)
	}

	// Step 2: Fetch the created records using GetRecordsByIds
	fmt.Println("\nFetching accounts by IDs...")

	res, err := conn.GetRecordsByIds(ctx, common.ReadByIdsParams{
		ObjectName: "accounts",
		RecordIds:  recordIDs,
		Fields:     []string{"id", "name", "domain"},
	})
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

	fmt.Println("\n✓ Test completed successfully!")
	fmt.Println("✓ All test accounts have been cleaned up.")
}
