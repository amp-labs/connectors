package front

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "empty.json")
	unsupportedResponse := testutils.DataFromFile(t, "notfound.json")
	contacts := testutils.DataFromFile(t, "contacts.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "meme", Fields: datautils.NewStringSet("name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "contact_lists", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Contacts",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, contacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "crd_41fx7bw",
						"name": "Willy Hernandez",
					},
					Raw: map[string]any{
						"_links": map[string]any{
							"self": "https://ampersand.api.frontapp.com/contacts/crd_41fx7bw",
							"related": map[string]any{
								"notes":         "https://ampersand.api.frontapp.com/contacts/crd_41fx7bw/notes",
								"conversations": "https://ampersand.api.frontapp.com/contacts/crd_41fx7bw/conversations",
								"owner":         nil,
							},
						},
						"id":          "crd_41fx7bw",
						"name":        "Willy Hernandez",
						"description": "",
						"updated_at":  1715626085.049,
						"is_private":  false,
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
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
