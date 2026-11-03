package grain

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseUploadCreate := testutils.DataFromFile(t, "write/upload-create.json")
	responseUploadUpdate := testutils.DataFromFile(t, "write/upload-update.json")
	responseBadRequest := testutils.DataFromFile(t, "write/error-bad-request.json")

	apiVersion := mockcond.Header(http.Header{"Public-Api-Version": []string{"2026-10-01"}})

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "uploads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Read-only object is not writable",
			Input: common.WriteParams{
				ObjectName: "teams",
				RecordId:   "9d228ffa-e5f7-4dd7-ab0a-5169fb878e21",
				RecordData: map[string]any{"name": "Sales"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Recordings cannot be created",
			Input:        common.WriteParams{ObjectName: "recordings", RecordData: map[string]any{"title": "Sync"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Recording title is updated with the API version",
			Input: common.WriteParams{
				ObjectName: "recordings",
				RecordId:   "0470708a-df31-4386-9a66-f8b2ae480f70",
				RecordData: map[string]any{"title": "Ampersand proxy test (updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/_/public-api/v2/recordings/0470708a-df31-4386-9a66-f8b2ae480f70"),
					apiVersion,
					mockcond.Body(`{"title": "Ampersand proxy test (updated)"}`),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"success": true}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0470708a-df31-4386-9a66-f8b2ae480f70",
			},
		},
		{
			Name: "Upload is created at the create endpoint",
			Input: common.WriteParams{
				ObjectName: "uploads",
				RecordData: map[string]any{
					"media_type":   "transcript",
					"auto_promote": false,
					"attributes": map[string]any{
						"title": "Ampersand schema validation (staged, do not promote)",
						"transcript": map[string]any{
							"language": "en",
							"segments": []any{
								map[string]any{"speaker_key": "s1", "start": 0, "end": 1200, "text": "Schema validation."},
							},
						},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/_/public-api/v2/uploads/create"),
					apiVersion,
					mockcond.Body(`{
						"media_type": "transcript",
						"auto_promote": false,
						"attributes": {
							"title": "Ampersand schema validation (staged, do not promote)",
							"transcript": {
								"language": "en",
								"segments": [{"speaker_key": "s1", "start": 0, "end": 1200, "text": "Schema validation."}]
							}
						}
					}`),
				},
				Then: mockserver.Response(http.StatusCreated, responseUploadCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0efac057-7d26-4d82-afc7-771fc081aa51",
				Data: map[string]any{
					"id":           "0efac057-7d26-4d82-afc7-771fc081aa51",
					"status":       "staged",
					"auto_promote": false,
					"recording_id": nil,
				},
			},
		},
		{
			Name: "Staged upload attributes are updated",
			Input: common.WriteParams{
				ObjectName: "uploads",
				RecordId:   "0efac057-7d26-4d82-afc7-771fc081aa51",
				RecordData: map[string]any{
					"attributes": map[string]any{"title": "Ampersand schema validation (staged, updated)"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/_/public-api/v2/uploads/0efac057-7d26-4d82-afc7-771fc081aa51"),
					apiVersion,
					mockcond.Body(`{"attributes": {"title": "Ampersand schema validation (staged, updated)"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseUploadUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0efac057-7d26-4d82-afc7-771fc081aa51",
				Data: map[string]any{
					"title":            "Ampersand schema validation (staged, updated)",
					"updated_datetime": "2026-10-28T11:37:48Z",
				},
			},
		},
		{
			// Grain rejects any property its endpoint does not define, such as a read-only field.
			Name: "Validation error is parsed",
			Input: common.WriteParams{
				ObjectName: "recordings",
				RecordId:   "0470708a-df31-4386-9a66-f8b2ae480f70",
				RecordData: map[string]any{"title": "Ampersand proxy test", "share_state": "public"},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Schema does not allow additional properties. At path: #/share_state."),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
