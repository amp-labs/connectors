package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/apollo"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	err := testCreatingOpportunities(context.Background())
	if err != nil {
		return 1
	}

	err = testUpdatingOpportunities(context.Background())
	if err != nil {
		return 1
	}

	return 0
}

func testCreatingOpportunities(ctx context.Context) error {
	conn := apollo.GetApolloConnector(ctx, "apollo-creds.json")

	params := common.WriteParams{
		ObjectName: "opportunities",
		RecordData: map[string]any{
			"name":                 "opportunity - one",
			"amount":               "200",
			"opportunity_stage_id": "65b1974393794c0300d26dcf",
			"closed_date":          "2024-12-18",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
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

func testUpdatingOpportunities(ctx context.Context) error {
	conn := apollo.GetApolloConnector(ctx, "apollo-creds.json")

	params := common.WriteParams{
		ObjectName: "opportunities",
		RecordId:   "66d19d6f0cb92801b3027306",
		RecordData: map[string]any{
			"amount":               "250",
			"opportunity_stage_id": "65b1974393794c0300d26dcf",
			"closed_date":          "2024-12-18",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
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

func testCreatingAccounts(ctx context.Context) error {
	conn := apollo.GetApolloConnector(ctx, "apollo-creds.json")

	params := common.WriteParams{
		ObjectName: "accounts",
		RecordData: map[string]any{
			"name":         "Google",
			"domain":       "google.com",
			"phone_number": "1-866-246-6453",
			"raw_address":  "1600 Amphitheatre Parkway",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		log.Fatal(err.Error())
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
