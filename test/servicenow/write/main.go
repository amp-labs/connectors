package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
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

	err = testCreateCase(ctx, conn)
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
			"first_name":        "Mohammed",
			"last_name":         "Salah",
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
	// Fetch a real lead id from the instance rather than hardcoding one, so the
	// update targets an existing record.
	read, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "lead",
		Fields:     datautils.NewStringSet("sys_id"),
	})
	if err != nil {
		return fmt.Errorf("reading a lead to update: %w", err)
	}

	if len(read.Data) == 0 {
		return fmt.Errorf("no lead found to update")
	}

	recordID, _ := read.Data[0].Fields["sys_id"].(string)

	params := common.WriteParams{
		ObjectName: "lead",
		RecordId:   recordID,
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

func testCreateCase(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "case",
		RecordData: map[string]any{
			"short_description": "Customer reported a billing discrepancy",
			"priority":          "2",
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
