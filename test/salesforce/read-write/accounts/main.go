package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

const TimeoutSeconds = 30

func main() {
	os.Exit(mainFn())
}

func mainFn() int { //nolint:funlen
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)

	if err := testReadConnector(ctx, conn); err != nil {
		slog.Error("Error testing", "connector", conn, "error", err)

		return 1
	}

	if err := testWriteConnector(ctx, conn); err != nil {
		slog.Error("Error testing", "connector", conn, "error", err)

		return 1
	}

	return 0
}

func testReadConnector(ctx context.Context, conn *salesforce.Connector) error {
	// Create a context with a timeout
	ctx, done := context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	// Read some data from Salesforce
	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Account",
		Fields:     connectors.Fields("Id", "Name", "BillingCity", "IsDeleted"),
	})
	if err != nil {
		return fmt.Errorf("error reading from Salesforce: %w", err)
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testWriteConnector(ctx context.Context, conn *salesforce.Connector) error {
	// IMPORTANT: every time this test is run, it will create a new Account
	// in SFDC instance. Will need to delete those out at later date.
	writtenRecordId, err := testSalesforceValidCreate(ctx, conn)
	if err != nil {
		return fmt.Errorf("error creating record in Salesforce: %w", err)
	}
	// IMPORTANT: will fail if specific recordId does not already exist in instance
	if err := testSalesforceValidUpdate(ctx, conn, writtenRecordId); err != nil {
		return fmt.Errorf("error updating record in Salesforce: %w", err)
	}

	return nil
}

const accountNumber = 123

// testSalesforceValidCreate will create a valid record in Salesforce.
func testSalesforceValidCreate(ctx context.Context, conn *salesforce.Connector) (string, error) {
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Account",
		RecordData: map[string]any{
			"Name":          "TEST ACCOUNT - [TO DELETE]",
			"AccountNumber": accountNumber,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error writing to Salesforce: %w", err)
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(writeRes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return writeRes.RecordId, nil
}

const accountNumber2 = 456

// testSalesforceValidUpdate will update existing record in Salesforce.
func testSalesforceValidUpdate(ctx context.Context, conn *salesforce.Connector, writtenRecordId string) error {
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Account",
		RecordData: map[string]any{
			"Name":          "OKADA TEST ACCOUNT",
			"AccountNumber": accountNumber2,
		},
		RecordId: writtenRecordId,
	})
	if err != nil {
		return fmt.Errorf("error writing to Salesforce: %w", err)
	}

	if !writeRes.Success {
		return fmt.Errorf("write to %s failed when it should have succeeded", writtenRecordId) //nolint:goerr113
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(writeRes, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
