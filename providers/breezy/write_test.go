package breezy

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createResp := testutils.DataFromFile(t, "write/position-create.json")
	updateResp := testutils.DataFromFile(t, "write/position-update.json")

	tests := []testroutines.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectPositions},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.WriteParams{ObjectName: "templates", RecordData: map[string]any{"name": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create position",
			Input: common.WriteParams{
				ObjectName: objectPositions,
				RecordData: map[string]any{
					"name":        "Proxy Test Position",
					"type":        "fullTime",
					"description": "Created via connector test",
					"location": map[string]any{
						"country":   "US",
						"state":     "CA",
						"city":      "San Francisco",
						"is_remote": true,
					},
					"department": "Engineering",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/company/" + testCompanyID + "/positions"),
					mockcond.Body(`{"department":"Engineering","description":"Created via connector test","location":{"city":"San Francisco","country":"US","is_remote":true,"state":"CA"},"name":"Proxy Test Position","type":"fullTime"}`),
				},
				Then: mockserver.Response(http.StatusOK, createResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "pos_new",
				Data: map[string]any{
					"_id":         "pos_new",
					"name":        "Proxy Test Position",
					"type":        "fullTime",
					"state":       "draft",
					"description": "Created via connector test",
					"department":  "Engineering",
				},
			},
		},
		{
			Name: "Update position",
			Input: common.WriteParams{
				ObjectName: objectPositions,
				RecordId:   "pos001",
				RecordData: map[string]any{
					"name":        "Software Engineer (Updated)",
					"type":        "fullTime",
					"description": "Updated via connector test",
					"department":  "Engineering",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3/company/" + testCompanyID + "/position/pos001"),
					mockcond.Body(`{"department":"Engineering","description":"Updated via connector test","name":"Software Engineer (Updated)","type":"fullTime"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateResp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "pos001",
				Data: map[string]any{
					"_id":         "pos001",
					"name":        "Software Engineer (Updated)",
					"type":        "fullTime",
					"state":       "published",
					"description": "Updated via connector test",
					"department":  "Engineering",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
