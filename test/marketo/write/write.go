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

	//Uses standard ids
	err := testWriteLeads(ctx)
	if err != nil {
		return 1
	}

	// Uses marketoGUIDs
	err = testWriteOpportunities(ctx)
	if err != nil {
		return 1
	}

	err = testUpdateLeads(ctx)
	if err != nil {
		return 1
	}

	err = testcreateCustomLeadField(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func testWriteLeads(ctx context.Context) error {
	conn := marketo.GetMarketoConnectorLeads(ctx)

	params := common.WriteParams{
		ObjectName: "leads",
		RecordData: map[string]any{
			"email":     gofakeit.Email(),
			"firstName": "Example Lead",
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
	conn := marketo.GetMarketoConnectorLeads(ctx)

	params := common.WriteParams{
		ObjectName: "opportunities",
		RecordData: map[string]any{
			"externalopportunityid": gofakeit.RandomString([]string{"opportunity 01", "opportunity 02", "opportunity 03", "opportunity 04"}),
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error(err.Error())
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

func testUpdateLeads(ctx context.Context) error {
	conn := marketo.GetMarketoConnectorLeads(ctx)

	params := common.WriteParams{
		ObjectName: "leads",
		RecordId:   "576",
		RecordData: map[string]any{
			"email": "babaknows@example.com",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error(err.Error())
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

func testcreateCustomLeadField(ctx context.Context) error {
	conn := marketo.GetMarketoConnectorLeads(ctx)

	params := common.WriteParams{
		ObjectName: "leads/schema/fields",
		RecordData: map[string]any{
			"displayName": "active",
			"name":        "active",
			"dataType":    "boolean",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Error(err.Error())
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
