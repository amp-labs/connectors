package outplay

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createProspectResponse := testutils.DataFromFile(t, "create-prospect.json")
	createProspectAccountResponse := testutils.DataFromFile(t, "create-prospectaccount.json")
	createNoteResponse := testutils.DataFromFile(t, "create-note.json")
	updateNoteResponse := testutils.DataFromFile(t, "update-note.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "prospect"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Successfully create a prospect",
			Input: common.WriteParams{
				ObjectName: "prospect",
				RecordData: map[string]any{
					"emailid":   "john.doe@example.com",
					"firstname": "John",
					"lastname":  "Doe",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v1/prospect"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createProspectResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "784158",
				Data: map[string]any{
					"prospectid":  float64(784158),
					"emailid":     "timezone2343@martinz.co.in",
					"firstname":   "time",
					"lastname":    "zone234",
					"phone":       "+919959626133",
					"designation": "Sales Executive12",
					"timezone":    "America/Los_Angeles",
					"city":        "hyderabad",
					"linkedin":    "https://linkedin.com/suresh_ana",
					"state":       "tg",
					"country":     "india",
					"twitter":     "twitter.com",
					"facebook":    "fb.om",
					"company":     "oracle",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully create a prospect account",
			Input: common.WriteParams{
				ObjectName: "prospectaccount",
				RecordData: map[string]any{
					"name":        "audition",
					"externalid":  "2",
					"description": "A test company",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v1/prospectaccount"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createProspectAccountResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "17718",
				Data: map[string]any{
					"accountid":     float64(17718),
					"name":          "audition",
					"externalid":    "2",
					"description":   "A test company",
					"employeecount": float64(122),
					"industrytype":  "Sales",
					"linkedin":      "http://linkedin.com/suresh_anaparthi",
					"twitter":       "http://twitter.com/suresh_anaparthi",
					"foundedyear":   "2002",
					"city":          "hyderabad",
					"website":       "",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully create a note",
			Input: common.WriteParams{
				ObjectName: "note",
				RecordData: map[string]any{
					"title":   "Meeting Notes",
					"content": "Discussed project requirements",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v1/note/create"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, createNoteResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "106128",
				Data: map[string]any{
					"success": true,
					"message": "Note saved successfully.",
					"noteId":  float64(106128),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update note with special update endpoint",
			Input: common.WriteParams{
				ObjectName: "note",
				RecordId:   "98765",
				RecordData: map[string]any{
					"title":   "Updated Meeting Notes",
					"content": "Updated project requirements discussion",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v1/note/update/98765"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, updateNoteResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"success":    true,
					"statuscode": float64(0),
					"message":    "Note saved successfully.",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
