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

	err := testWriteIncidents(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateIncidents(ctx, conn)
	if err != nil {
		return err
	}

	err = testUpdateMailServer(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testWriteIncidents(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "incident",
		RecordData: map[string]any{
			"assigned_to": "1c741bd70b2322007518478d83673af3",
			"urgency":     "1",
			"comments":    "Elevating urgency, this is a blocking issue",
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

func testUpdateIncidents(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "incident",
		RecordId:   "6a82f62b836f5610e945a6d0deaad369",
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

func testUpdateMailServer(ctx context.Context, conn *ServiceNow.Connector) error {
	params := common.WriteParams{
		ObjectName: "cmdb_ci_email_server",
		RecordId:   "280ffff1c0a8000b0083f5395b44bc97",
		RecordData: map[string]any{
			"due_in": "2025-10-10",
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
