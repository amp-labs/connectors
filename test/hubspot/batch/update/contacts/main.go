package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetHubspotConnector(ctx)

	tests := []struct {
		name    string
		records common.BatchItems
	}{
		{
			name: "Duplicate identifiers for both records",
			records: common.BatchItems{{
				Record: map[string]any{
					"id":        "123456",
					"lastname":  "Dyer (updated)",
					"firstname": "Siena (updated)",
				},
			}, {
				Record: map[string]any{
					"id":        "123456",
					"lastname":  "Blevins (updated)",
					"firstname": "Markus (updated)",
				},
			}},
		},
		{
			name: "Missing identifier for one record",
			records: common.BatchItems{{
				Record: map[string]any{
					"id":        "171591000199",
					"lastname":  "Dyer (updated)",
					"firstname": "Siena (updated)",
				},
			}, {
				Record: map[string]any{
					"lastname":  "Blevins (updated)",
					"firstname": "Markus (updated)",
				},
			}},
		},
		{
			name: "Invalid fields but has identifiers",
			records: common.BatchItems{{
				Record: map[string]any{
					"id":           "171591000199",
					"last000name":  "Dyer (updated)",
					"first000name": "Siena (updated)",
				},
			}, {
				Record: map[string]any{
					"id":           "171591000198",
					"last000name":  "Blevins (updated)",
					"first000name": "Markus (updated)",
				},
			}},
		},
		{
			name: "Both records are valid",
			records: common.BatchItems{{
				Record: map[string]any{
					"id":        "171591000199",
					"lastname":  "Dyer (updated)",
					"firstname": "Siena (updated)",
				},
			}, {
				Record: map[string]any{
					"id":        "171591000198",
					"lastname":  "Blevins (updated)",
					"firstname": "Markus (updated)",
				},
			}},
		},
	}

	for _, tt := range tests {
		res, err := conn.BatchWrite(ctx, &connectors.BatchWriteParam{
			ObjectName: "contacts",
			Type:       connectors.BatchWriteTypeUpdate,
			Batch:      tt.records,
		})
		if err != nil {
			utils.Fail("error reading", "error", err)
		}

		fmt.Println(tt.name)
		utils.DumpJSON(res, os.Stdout)
	}
}
