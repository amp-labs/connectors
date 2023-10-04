package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/salesforce"
	"golang.org/x/oauth2"
)

// To run this test, first generate a Salesforce Access token
// (https://ampersand.slab.com/posts/salesforce-api-guide-go1d9wnj#h0ciq-generate-an-access-token)

// Then set the appropriate oauth fields below, and then run.
// go run test/salesforce.go

// You can optionally add an `instance` argument to specify a Salesforce instance,
// or leave empty to use the Ampersand's dev instance.

const TimeoutSeconds = 30

func main() {
	os.Exit(mainFn())
}

func mainFn() int {
	subdomain := flag.String("subdomain", "boxit2-dev-ed", "Salesforce subdomain")
	flag.Parse()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cfg := &oauth2.Config{
		ClientID:     "3MVG9kBt168mda__AsLfwj2vUtrPMp39Nvj9amL1F7wMQhoDK7FgznCLTvYMIYLcDidAVGom5YCeiVbbFkE3X",
		ClientSecret: "E1B56598590E8189149988F042160B28AA7771F9EDC3DDB2650263382BD9EDEB",
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", *subdomain),
			TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", *subdomain),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	tok := &oauth2.Token{
		AccessToken:  "00D4x000006ll1H!AQMAQFsLsN.FGi6010dYa3SLi6S2dljLTGOHg_5Q83i3qAF18HWlHbzOyvqMbIg4mzW_GkAZKgsw0Xt6ysBhKDs5us58NaJ9",
		RefreshToken: "5Aep861Bky2w54txC2OK9jDRPrEZfc98zscX2XARXYBQSBJb61Ksw3TagOU9iKqqJmZs293Eqwszajmegv9XD3I",
		TokenType:    "bearer",
		Expiry:       time.Now().Add(-1 * time.Hour), // just pretend it's expired already, whatever, it'll fetch a new one.
	}

	ctx := context.Background()

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	sfc, err := connectors.Salesforce.New(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithSubdomain(*subdomain))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return 1
	}

	defer func() {
		_ = sfc.Close()
	}()

	if err := testConnector(ctx, sfc); err != nil {
		slog.Error("Error testing", "connector", sfc, "error", err)

		return 1
	}

	return 0
}

func testConnector(ctx context.Context, conn connectors.Connector) error {
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

	// IMPORTANT: every time this test is run, it will create a new Account
	// in SFDC instance. Will need to delete those out at later date.
	writtenObjId, err := testSalesforceValidCreate(ctx, conn)
	if err != nil {
		return fmt.Errorf("error creating record in Salesforce: %w", err)
	}
	// IMPORTANT: will fail if specific objectId does not already exist in instance
	if err := testSalesforceValidUpdate(ctx, conn, writtenObjId); err != nil {
		return fmt.Errorf("error updating record in Salesforce: %w", err)
	}

	return nil
}

// Create a valid record in Salesforce
func testSalesforceValidCreate(ctx context.Context, conn connectors.Connector) (string, error) {

	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Account",
		ObjectData: map[string]interface{}{
			"Name":          "TEST ACCOUNT - [TO DELETE]",
			"AccountNumber": 123,
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

	return writeRes.ObjectId, nil
}

// Update existing record in Salesforce
func testSalesforceValidUpdate(ctx context.Context, conn connectors.Connector, writtenObjId string) error {
	writeRes, err := conn.Write(ctx, connectors.WriteParams{
		ObjectName: "Account",
		ObjectData: map[string]interface{}{
			"Name":          "OKADA TEST ACCOUNT",
			"AccountNumber": 456,
		},
		ObjectId: writtenObjId,
	})
	if err != nil {
		return fmt.Errorf("error writing to Salesforce: %w", err)
	}
	if !writeRes.Success {
		return fmt.Errorf("write to %s failed when it should have succeeded", writtenObjId)
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
