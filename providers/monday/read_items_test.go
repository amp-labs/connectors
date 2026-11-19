package monday

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestReadItems(t *testing.T) {
	t.Parallel()

	itemsResponse := `{
		"data": {
			"boards": [{
				"items_page": {
					"cursor": "next-cursor",
					"items": [{
						"id": "9876543210",
						"name": "Item 1",
						"column_values": [
							{"id": "status", "text": "Done", "type": "status"},
							{"id": "text", "text": "hello", "type": "text"}
						]
					}]
				}
			}]
		}
	}`

	tests := []testroutines.Read{
		{
			Name: "Read items requires board_id",
			Input: common.ReadParams{
				ObjectName: mondayObjectItems,
				Fields:     connectors.Fields("id", "name"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrBoardIDRequired},
		},
		{
			Name: "Read items flattens cf_ column values",
			Input: common.ReadParams{
				ObjectName: mondayObjectItems,
				Fields:     connectors.Fields("id", "cf_status", "cf_text"),
				Filter:     "board_id=1234567890",
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, []byte(itemsResponse)),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":        "9876543210",
						"cf_status": "Done",
						"cf_text":   "hello",
					},
					Raw: map[string]any{
						"id":   "9876543210",
						"name": "Item 1",
						"column_values": []any{
							map[string]any{"id": "status", "text": "Done", "type": "status"},
							map[string]any{"id": "text", "text": "hello", "type": "text"},
						},
					},
				}},
				Done: true,
			},
			Comparator: testroutines.ComparatorSubsetRead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
