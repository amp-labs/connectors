package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	ServiceNow "github.com/amp-labs/connectors/providers/servicenow"
	"github.com/amp-labs/connectors/test/servicenow"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := servicenow.GetServiceNowConnector(ctx)

	err := testCreateLead(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateLead(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateCase(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreateContact(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreateLead(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "lead",
		RecordData: map[string]any{
			"short_description": "Interested in premium plan",
			"state":             "1",
			"company":           "Withampersand",
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

func testUpdateLead(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "lead",
		RecordId:   "6a2f6fbb83f02210290fed70deaad320",
		RecordData: map[string]any{
			"company": "Ampersand",
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

func testUpdateCase(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "case",
		RecordId:   "280ffff1c0a8000b0083f5395b44bc97",
		RecordData: map[string]any{
			"priority": "2",
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

func testCreateContact(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "contact",
		RecordData: map[string]any{
			"active":       true,
			"agent_status": "On break",
			"city":         "Liverpool",
			"company":      "Withampersand",
			"country":      "UK",
			"email":        "ywnwa@lfc.com",
			"first_name":   "Mohammed",
			"last_name":    "Salah",
			"phone":        "+17689673546",
			"gender":       "male",
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
