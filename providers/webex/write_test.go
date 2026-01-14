package webex

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint:funlen,gocognit,cyclop
func TestWrite(t *testing.T) {
	t.Parallel()

	responseCreateUpdatePerson := testutils.DataFromFile(t, "create-update-person.json")
	responseUpdateGroup := testutils.DataFromFile(t, "write-group.json")
	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "people"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name: "Write must act as a Create (POST)",
			Input: common.WriteParams{
				ObjectName: "people",
				RecordData: map[string]any{
					"emails":      []any{"exmple@example.com"},
					"displayName": "Example Person",
					"firstName":   "Example",
					"lastName":    "Person",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateUpdatePerson),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "Y2lzY29zcGFyazovL3VzL1BFT1BMRS9",
				Errors:   nil,
				Data: map[string]any{
					"id":          "Y2lzY29zcGFyazovL3VzL1BFT1BMRS9",
					"displayName": "Example Person",
					"firstName":   "Example",
					"lastName":    "Person",
					"emails":      []any{"exmple@example.com"},
					"type":        "person",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Write must act as an Update (PUT)",
			Input: common.WriteParams{
				ObjectName: "people",
				RecordId:   "Y2lzY29zcGFyazovL3VzL1BFT1BMRS9",
				RecordData: map[string]any{
					"emails":      []any{"exmple.updated@example.com"},
					"displayName": "Example Person Updated",
					"firstName":   "Example",
					"lastName":    "Person",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v1/people/Y2lzY29zcGFyazovL3VzL1BFT1BMRS9"),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateUpdatePerson),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "Y2lzY29zcGFyazovL3VzL1BFT1BMRS9",
				Errors:   nil,
				Data: map[string]any{
					"id":          "Y2lzY29zcGFyazovL3VzL1BFT1BMRS9",
					"displayName": "Example Person",
					"firstName":   "Example",
					"lastName":    "Person",
					"emails":      []any{"exmple@example.com"},
					"type":        "person",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Write groups must act as an Update (PATCH)",
			Input: common.WriteParams{
				ObjectName: "groups",
				RecordId:   "Y2lzY29zcyZj",
				RecordData: map[string]any{
					"displayName": "Updated Group Name",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1/groups/Y2lzY29zcyZj"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateGroup),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "Y2lzY29zcyZj",
				Errors:   nil,
				Data: map[string]any{
					"id":          "Y2lzY29zcyZj",
					"displayName": "Site1",
				},
			},
			ExpectedErrs: nil,
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
