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
	ctx := context.Background()

	err := testCreatingOpportunities(ctx)
	if err != nil {
		return 1
	}

	err = testUpdatingOpportunities(ctx)
	if err != nil {
		return 1
	}

	err = testCreatingAccounts(ctx)
	if err != nil {
		return 1
	}

	err = testUpdatingDeals(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testCreatingOpportunities(ctx context.Context) error {
	conn := apollo.GetApolloConnector(ctx)

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
	conn := apollo.GetApolloConnector(ctx)

	params := common.WriteParams{
		ObjectName: "opportunities",
		RecordId:   "66d573f1bb530101b230db6f",
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
	conn := apollo.GetApolloConnector(ctx)

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

func testUpdatingDeals(ctx context.Context) error {
	conn := apollo.GetApolloConnector(ctx)

	params := common.WriteParams{
		ObjectName: "Deals",
		RecordId:   "66d573f1bb530101b230db6f",
		RecordData: map[string]any{
			"amount":               "2500",
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
