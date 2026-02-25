package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/revenuecat"
	connTest "github.com/amp-labs/connectors/test/revenuecat"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetRevenueCatConnector(ctx)

	// Customers cannot be created via the REST API (created by the mobile SDK).
	// This test updates and deletes an existing customer.
	slog.Info("=== customers (update -> delete) ===")

	customerID, err := getFirstCustomerID(ctx, conn)
	if err != nil {
		slog.Error("Failed to read customers (needed for customer_id)", "error", err)
		return 1
	}
	slog.Info("Using customer", "id", customerID)

	if err := updateCustomer(ctx, conn, customerID); err != nil {
		slog.Error("Failed to update customer", "error", err, "customer_id", customerID)
		return 1
	}

	if err := deleteByID(ctx, conn, "customers", customerID); err != nil {
		slog.Error("Failed to delete customer", "error", err, "customer_id", customerID)
		return 1
	}

	slog.Info("RevenueCat customers write-delete test completed successfully")
	return 0
}

func getFirstCustomerID(ctx context.Context, conn *revenuecat.Connector) (string, error) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "customers",
		Fields:     connectors.Fields("id"),
		PageSize:   1,
	})
	if err != nil {
		return "", err
	}

	if len(res.Data) == 0 {
		return "", fmt.Errorf("no customers found in project")
	}

	id, _ := res.Data[0].Raw["id"].(string)
	if id == "" {
		return "", fmt.Errorf("customer response missing id field")
	}

	return id, nil
}

func updateCustomer(ctx context.Context, conn *revenuecat.Connector, customerID string) error {
	slog.Info("Updating customer attributes", "customer_id", customerID)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "customers",
		RecordId:   customerID,
		RecordData: map[string]any{
			"attributes": map[string]any{
				"$displayName": map[string]any{
					"value": "Amp WD Test Customer",
				},
			},
		},
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)
	return nil
}

func deleteByID(ctx context.Context, conn *revenuecat.Connector, objectName, recordID string) error {
	slog.Info("Deleting record", "object", objectName, "id", recordID)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   recordID,
	})
	if err != nil {
		return err
	}
	utils.DumpJSON(res, os.Stdout)

	if !res.Success {
		return fmt.Errorf("delete reported Success=false")
	}
	return nil
}
