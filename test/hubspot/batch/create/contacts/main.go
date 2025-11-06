package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

type record struct {
	Properties map[string]any `json:"properties"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	tests := []struct {
		name    string
		records []any
	}{
		{
			name: "Both records use conflicting emails",
			records: []any{
				record{Properties: map[string]any{
					"email":     "MarleyFleming@hubspot.com",
					"lastname":  "Dyer",
					"firstname": "Siena",
				}},
				record{Properties: map[string]any{
					"email":     "MarleyFleming@hubspot.com",
					"lastname":  "Blevins",
					"firstname": "Markus",
				}},
			},
		},
		{
			name: "One record uses unknown field",
			records: []any{
				record{Properties: map[string]any{
					"lastname":  "Dyer003",
					"firstname": "Siena003",
				}},
				record{Properties: map[string]any{
					"last5555name": "Blevins003",
					"firstname":    "Markus003",
				}},
			},
		},
		{
			name: "Both records are valid",
			records: []any{
				record{Properties: map[string]any{
					"lastname":  "Dyer",
					"firstname": "Siena",
				}},
				record{Properties: map[string]any{
					"lastname":  "Blevins",
					"firstname": "Markus",
				}},
			},
		},
	}

	for _, tt := range tests {
		res, err := conn.BatchWrite(ctx, &connectors.BatchWriteParam{
			ObjectName: "contacts",
			Type:       connectors.BatchWriteTypeCreate,
			Records:    tt.records,
		})
		if err != nil {
			utils.Fail("error reading", "error", err)
		}

		fmt.Println(tt.name)
		utils.DumpJSON(res, os.Stdout)
	}
}
