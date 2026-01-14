package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	cl "github.com/amp-labs/connectors/providers/calendly"
	"github.com/amp-labs/connectors/test/calendly"
)

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
	}
}

func run() error {
	ctx := context.Background()
	conn := calendly.GetCalendlyConnector(ctx)

	err := testCreatingSchedulingLink(ctx, conn)
	if err != nil {
		return err
	}

	err = testCreatingEventTypes(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

func testCreatingSchedulingLink(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "scheduling_links",
		RecordData: map[string]any{
			"max_event_count": 1,
			"owner":           "https://api.calendly.com/event_types/4a2abc24-beca-487f-8bcc-dcdbc20cb370",
			"owner_type":      "EventType",
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

func testCreatingEventTypes(ctx context.Context, conn *cl.Connector) error {
	params := common.WriteParams{
		ObjectName: "one_off_event_types",
		RecordData: map[string]any{
			"name":     "My Meeting Test 1",
			"host":     "https://api.calendly.com/users/42687819-a60c-446a-b42f-0d84ce589f0e",
			"duration": 30,
			"timezone": "Africa/Dar_es_Salaam",
			"date_setting": map[string]string{
				"type":       "date_range",
				"start_date": "2025-09-07",
				"end_date":   "2025-10-07",
			},
			"location": map[string]string{
				"kind":            "physical",
				"location":        "Main Office",
				"additional_info": "string",
			},
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
