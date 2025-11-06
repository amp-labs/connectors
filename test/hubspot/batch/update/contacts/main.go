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
	ID         string         `json:"id"`
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
			name: "Duplicate identifiers for both records",
			records: []any{
				record{
					ID: "123456",
					Properties: map[string]any{
						"lastname":  "Dyer (updated)",
						"firstname": "Siena (updated)",
					},
				},
				record{
					ID: "123456",
					Properties: map[string]any{
						"lastname":  "Blevins (updated)",
						"firstname": "Markus (updated)",
					},
				},
			},
		},
		{
			name: "Missing identifier for one record",
			records: []any{
				record{
					ID: "171591000199",
					Properties: map[string]any{
						"lastname":  "Dyer (updated)",
						"firstname": "Siena (updated)",
					},
				},
				record{Properties: map[string]any{
					"lastname":  "Blevins (updated)",
					"firstname": "Markus (updated)",
				}},
			},
		},
		{
			name: "Invalid fields but has identifiers",
			records: []any{
				record{
					ID: "171591000199",
					Properties: map[string]any{
						"last000name":  "Dyer (updated)",
						"first000name": "Siena (updated)",
					},
				},
				record{
					ID: "171591000198",
					Properties: map[string]any{
						"last000name":  "Blevins (updated)",
						"first000name": "Markus (updated)",
					},
				},
			},
		},
		{
			name: "Both records are valid",
			records: []any{
				record{
					ID: "171591000199",
					Properties: map[string]any{
						"lastname":  "Dyer (updated)",
						"firstname": "Siena (updated)",
					},
				},
				record{
					ID: "171591000198",
					Properties: map[string]any{
						"lastname":  "Blevins (updated)",
						"firstname": "Markus (updated)",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		res, err := conn.BatchWrite(ctx, &connectors.BatchWriteParam{
			ObjectName: "contacts",
			Type:       connectors.BatchWriteTypeUpdate,
			Records:    tt.records,
		})
		if err != nil {
			utils.Fail("error reading", "error", err)
		}

		fmt.Println(tt.name)
		utils.DumpJSON(res, os.Stdout)
	}
}
