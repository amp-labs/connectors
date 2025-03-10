package main

// {

//   }

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cl "github.com/amp-labs/connectors/providers/clari"
	"github.com/amp-labs/connectors/test/clari"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := clari.GetConnector(ctx)

	err := testCreatingAuditEvents(ctx, conn)
	if err != nil {
		return err
	}

	err = testCancelAJob(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingAuditEvents(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "export/audit/events",
		RecordData: map[string]any{
			"actorId":              1,
			"impersonatingActorId": 1,
			"sessionId":            "fe795746-54b0-11ed-bdc3-0242ac120002",
			"sessionType":          "WEB",
			"dateFrom":             "2025-02-15T16:22:18Z",
			"dateTo":               "2025-03-05T16:22:18Z",
			"event":                "api.accessed",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
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

func testCancelAJob(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "export/jobs",
		RecordId:   "67cee51eccbb14039ee5696e",
		RecordData: map[string]any{
			"type": "CANCEL",
		},
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
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
