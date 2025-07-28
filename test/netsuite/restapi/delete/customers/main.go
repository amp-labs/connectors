package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/netsuite"
	"github.com/amp-labs/connectors/test/utils"
)

const TimeoutSeconds = 60

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetNetsuiteRESTAPIConnector(ctx)

	ctx, done = context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	// First, create a test customer to delete
	testCustomer := map[string]any{
		"companyName": "DELETE TEST " + time.Now().Format("20060102-150405"),
		"email":       "delete-test@example.com",
		"phone":       "555-DELETE",
		"subsidiary":  "1",
	}

	slog.Info("Creating test customer for deletion..")
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "customer",
		RecordData: testCustomer,
	})
	if err != nil {
		utils.Fail("error creating test customer", "error", err)
	}

	if !writeRes.Success || writeRes.RecordId == "" {
		utils.Fail("failed to create test customer or no record ID returned")
	}

	slog.Info("Test customer created", "recordId", writeRes.RecordId)

	// Now delete the customer
	slog.Info("Deleting customer..", "recordId", writeRes.RecordId)
	deleteRes, err := conn.Delete(ctx, connectors.DeleteParams{
		ObjectName: "customer",
		RecordId:   writeRes.RecordId,
	})
	if err != nil {
		utils.Fail("error deleting customer from NetSuite REST API", "error", err)
	}

	fmt.Println("Delete operation completed:")
	utils.DumpJSON(deleteRes, os.Stdout)

	if deleteRes.Success {
		slog.Info("Customer successfully deleted", "recordId", writeRes.RecordId)
	} else {
		slog.Error("Customer deletion failed", "recordId", writeRes.RecordId)
	}
}
