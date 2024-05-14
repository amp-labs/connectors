package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/salesforce"
	testUtils "github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/utils"
	"github.com/joho/godotenv"
)

// Set the appropriate environment variables in a .env file, then run:
// go run test/salesforce.go

const TimeoutSeconds = 30

func main() {
	os.Exit(mainFn())
}

func mainFn() int { //nolint:funlen
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "error", err)

		return 1
	}

	salesforceEnvVarPrefix := "SALESFORCE_"
	credentialsRegistry := utils.NewCredentialsRegistry()

	envSchema := testUtils.EnvVarsReaders(salesforceEnvVarPrefix)
	credentialsRegistry.AddReaders(
		envSchema...,
	)

	salesforceWorkspace := credentialsRegistry.MustString(utils.WorkspaceRef)

	cfg := utils.SalesforceOAuthConfigFromRegistry(credentialsRegistry)
	tok := utils.SalesforceOauthTokenFromRegistry(credentialsRegistry)
	ctx := context.Background()

	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(salesforceWorkspace))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return 1
	}

	defer func() {
		_ = sfc.Close()
	}()

	if err := testReadConnector(ctx, sfc); err != nil {
		slog.Error("Error testing", "connector", sfc, "error", err)

		return 1
	}

	if err := testWriteConnector(ctx, sfc); err != nil {
		slog.Error("Error testing", "connector", sfc, "error", err)

		return 1
	}

	return 0
}

func testReadConnector(ctx context.Context, conn connectors.ReadConnector) error {
	// Create a context with a timeout
	ctx, done := context.WithTimeout(ctx, TimeoutSeconds*time.Second)
	defer done()

	// Read some data from Salesforce
	res, err := conn.Read(ctx, connectors.ReadParams{
		ObjectName: "Account",
		Fields:     []string{"Id", "Name", "BillingCity", "IsDeleted"},
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

func testWriteConnector(ctx context.Context, conn connectors.WriteConnector) error {
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
func testSalesforceValidCreate(ctx context.Context, conn connectors.WriteConnector) (string, error) {
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Account",
		RecordData: map[string]interface{}{
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
func testSalesforceValidUpdate(ctx context.Context, conn connectors.WriteConnector, writtenRecordId string) error {
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Account",
		RecordData: map[string]interface{}{
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
