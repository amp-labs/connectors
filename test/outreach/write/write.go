package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/outreach"
	connTest "github.com/amp-labs/connectors/test/outreach"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetOutreachConnector(ctx)

	err := testWriteConnector(context.Background(), conn)
	if err != nil {
		utils.Fail("error writing to Outreach", "error", err)
	}
}

func testWriteConnector(ctx context.Context, conn *outreach.Connector) error {
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

func testCreateEmailAddresses(ctx context.Context, conn *outreach.Connector) error {
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

func testCreateProspects(ctx context.Context, conn *outreach.Connector) error {
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

func testCreateMailing(ctx context.Context, conn *outreach.Connector) error {
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

func testCreateSequence(ctx context.Context, conn *outreach.Connector) error {
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

func testUpdateEmailAddress(ctx context.Context, conn *outreach.Connector) error {
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
