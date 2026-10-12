package monday

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWriteItems(t *testing.T) {
	t.Parallel()

	columnsResponse := `{
		"data": {
			"boards": [{
				"columns": [
					{"id": "status", "title": "Status", "type": "status", "settings_str": ""}
				]
			}]
		}
	}`

	createItemResponse := `{
		"data": {
			"create_item": {
				"id": "111"
			}
		}
	}`

	tests := []testroutines.Write{
		{
			Name: "Create item maps cf_ keys to column_values mutation",
			Input: common.WriteParams{
				ObjectName: mondayObjectItems,
				RecordData: map[string]any{
					"board_id":  "1234567890",
					"group_id":  "topics",
					"name":      "New item",
					"cf_status": "Working on it",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.Check(func(_ http.ResponseWriter, r *http.Request) bool {
							return requestBodyContains(r, "boards(ids")
						}),
						Then: mockserver.Response(http.StatusOK, []byte(columnsResponse)),
					},
					{
						If: mockcond.Check(func(_ http.ResponseWriter, r *http.Request) bool {
							return requestBodyContains(r, "create_item")
						}),
						Then: mockserver.Response(http.StatusOK, []byte(createItemResponse)),
					},
				},
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "111",
				Data: map[string]any{
					"data": map[string]any{
						"create_item": map[string]any{
							"id": "111",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetWrite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
