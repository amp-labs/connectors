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
	conn := calendly.GetConnector(ctx)

	// err := testCreatingSchedulingLink(ctx, conn)
	// if err != nil {
	// 	return err
	// }

	err := testCreatingEventTypes(ctx, conn)
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
			"owner":           "https://api.calendly.com/event_types/012345678901234567890",
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
			"host":     "https://api.calendly.com/users/8e1d1ec8-5b1f-4a9c-aa80-bbb47694ea46",
			"duration": 30,
			"timezone": "Africa/Dar_es_Salaam",
			"date_setting": map[string]string{
				"type":       "date_range",
				"start_date": "2025-04-07",
				"end_date":   "2025-04-07",
			},
			"location": map[string]string{
				"kind":           "physical",
				"location":       "Main Office",
				"additonal_info": "string",
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
