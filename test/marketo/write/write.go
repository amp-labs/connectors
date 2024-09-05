package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/marketo"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	os.Exit(MainFn())
}

func MainFn() int {
	ctx := context.Background()

	err := testWriteLeads(ctx)
	if err != nil {
		return 1
	}

	err = testWriteOpportunities(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testWriteLeads(ctx context.Context) error {
	conn := marketo.GetMarketoConnectorW(ctx)

	params := common.WriteParams{
		ObjectName: "leads",
		RecordData: map[string]any{
			"input": []map[string]any{
				{
					"email":     gofakeit.Email(),
					"firstName": "Example Lead",
				},
			},
			"action":      "createOnly",
			"lookupField": "email",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		fmt.Println("ERR: ", err)
		return err
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

func testWriteOpportunities(ctx context.Context) error {
	conn := marketo.GetMarketoConnectorW(ctx)

	params := common.WriteParams{
		ObjectName: "opportunities",
		RecordData: map[string]any{
			"input": []map[string]any{
				{
					"externalopportunityid": gofakeit.RandomString([]string{"opportunity 1", "opportunity 2", "opportunity 3", "opportunity 4"}),
				},
			},
			"action": "createOnly",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Info(err.Error())
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
