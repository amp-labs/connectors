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

	// Test data for creating a customer
	testCustomer := map[string]any{
		"companyName": "Test Company " + time.Now().Format("20060102-150405"),
		"email":       "test@example.com",
		"phone":       "555-0123",
		"subsidiary":  "1",
	}

	slog.Info("Creating customer..")
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "customer",
		RecordData: testCustomer,
	})
	if err != nil {
		utils.Fail("error creating customer", "error", err)
	}

	slog.Info("Customer created", "writeResult", writeRes)

	if writeRes.Success && writeRes.RecordId != "" {
		// Read the created customer back
		slog.Info("Reading back created customer..")
		readRes, err := conn.Read(ctx, connectors.ReadParams{
			ObjectName: "customer",
			Fields:     connectors.Fields("id", "companyName", "email", "phone", "entityStatus"),
			Since:      time.Now().Add(-1 * time.Minute),
		})
		if err != nil {
			utils.Fail("error reading customer", "error", err)
		}

		fmt.Println("Read result:")
		utils.DumpJSON(readRes, os.Stdout)

		// Update the customer
		slog.Info("Updating customer..")
		updateData := map[string]any{
			"phone": "555-9999",
			"email": "updated@example.com",
		}

		updateRes, err := conn.Write(ctx, connectors.WriteParams{
			ObjectName: "customer",
			RecordId:   writeRes.RecordId,
			RecordData: updateData,
		})
		if err != nil {
			utils.Fail("error updating customer", "error", err)
		}

		slog.Info("Customer updated", "updateResult", updateRes)

		// Optionally delete the test customer (commented out for safety)
		// slog.Info("Deleting test customer..")
		// deleteRes, err := conn.Delete(ctx, connectors.DeleteParams{
		// 	ObjectName: "customer",
		// 	RecordId:   writeRes.RecordId,
		// })
		// if err != nil {
		// 	utils.Fail("error deleting customer", "error", err)
		// }
		// slog.Info("Customer deleted", "deleteResult", deleteRes)
	}

	slog.Info("Read-write test completed successfully")
}
