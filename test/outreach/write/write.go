package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/outreach"
	"github.com/brianvoe/gofakeit/v6"
)

const (
	DefaultCredsFile = "creds.json"
)

func main() {
	var err error

	outreach := outreach.GetOutreachConnector(context.Background(), DefaultCredsFile)

	err = testWriteConnector(context.Background(), outreach)
	if err != nil {
		log.Fatal(err)
	}

}

func testWriteConnector(ctx context.Context, conn connectors.WriteConnector) error {
	var err error

	err = testCreateMailing(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateEmailAddresses(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateProspects(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateSequence(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateEmailAddress(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreateEmailAddresses(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "emailAddresses",
		RecordData: map[string]any{
			"email":     gofakeit.Email(),
			"emailType": "email",
			"order":     0,
			"status":    "null",
			"type":      "emailAddress",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreateProspects(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "prospects",
		RecordData: map[string]any{
			"accountName":    "Testing Account",
			"addressCity":    "SAN FRANCISCO",
			"addressCountry": "USA",
			"addressState":   "CA",
			"addressZip":     "0000000",
			"angelListUrl":   "https://withampersand.com",
			"campaignName":   "Joker's Marathon",
			"company":        "Ampersand",
			"websiteUrl1":    "https://withampersand.com",
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreateMailing(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "mailings",
		RecordData: map[string]any{
			"type":     "mailing",
			"bodyHtml": "<html><body><p>Here Goes your HTML email</p></body>></html>",
			"subject":  "string",
			"relationships": map[string]any{
				"mailbox": map[string]any{
					"data": map[string]any{
						"id":   1,
						"type": "mailbox",
					},
				},
				"prospect": map[string]any{
					"data": map[string]any{
						"id":   1,
						"type": "prospect",
					},
				},
			},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testCreateSequence(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "sequences",
		RecordData: map[string]any{
			"description": "A test sequence",
			"type":        "sequence",
			"name":        "string",
			"tags":        []string{"sequence", "coffee"},
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testUpdateEmailAddress(ctx context.Context, conn connectors.WriteConnector) error {
	config := common.WriteParams{
		ObjectName: "emailAddresses",
		RecordId:   "5",
		RecordData: map[string]any{
			"email": "groverstiedemann@lehner.io",
			"order": gofakeit.Number(0, 15),
		},
	}

	result, err := conn.Write(ctx, config)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
