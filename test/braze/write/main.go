package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	br "github.com/amp-labs/connectors/providers/braze"
	"github.com/amp-labs/connectors/test/braze"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()
	conn := braze.NewBrazeConnector(ctx)

	if err := createEmailTemplate(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := createCatalog(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	if err := updateEmailTemplate(ctx, conn); err != nil {
		slog.Error(err.Error())
	}

	return 0
}

func createEmailTemplate(ctx context.Context, conn *br.Connector) error {
	prms := common.WriteParams{
		ObjectName: "templates/email/create",
		RecordData: map[string]any{
			"template_name":     "welcome_email",
			"subject":           "Welcome to [Company Name]!",
			"body":              "<p>Hi {{first_name}},</p><p>Thank you for joining us! We're excited to have you.</p><p>Get started by exploring your account <a href='{{dashboard_url}}'>here</a>.</p><p>Cheers,<br>The {{company_name}} Team</p>",
			"plaintext_body":    "Hi {{first_name}},\n\nThank you for joining us! Get started here: {{dashboard_url}}\n\nCheers,\nThe {{company_name}} Team",
			"preheader":         "Start your journey with us today",
			"should_inline_css": true,
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func updateEmailTemplate(ctx context.Context, conn *br.Connector) error {
	prms := common.WriteParams{
		ObjectName: "templates/email/update",
		RecordData: map[string]any{
			"email_template_id": "7fa9d5e1-faac-47bd-ba8a-33827231740c",
			"should_inline_css": false,
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}

func createCatalog(ctx context.Context, conn *br.Connector) error {
	prms := common.WriteParams{
		ObjectName: "catalogs",
		RecordData: map[string]any{
			"name":        "a_restaurant",
			"description": "My Restaurants",
			"fields": []map[string]string{
				{
					"name": "id",
					"type": "string",
				},
				{
					"name": "Name",
					"type": "string",
				},
				{
					"name": "City",
					"type": "string",
				},
				{
					"name": "Cuisine",
					"type": "string",
				},
				{
					"name": "Rating",
					"type": "number",
				},
				{
					"name": "Loyalty_Program",
					"type": "boolean",
				},
				{
					"name": "Location",
					"type": "object",
				},
				{
					"name": "Top_Dishes",
					"type": "array",
				},
				{
					"name": "Created_At",
					"type": "time",
				},
			},
		},
	}

	result, err := conn.Write(ctx, prms)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonStr))

	return nil
}
