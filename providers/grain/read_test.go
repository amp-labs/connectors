package grain

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseRecordingsFirstPage := testutils.DataFromFile(t, "read/recordings-first-page.json")
	responseRecordingsLastPage := testutils.DataFromFile(t, "read/recordings-last-page.json")
	responseUploads := testutils.DataFromFile(t, "read/uploads.json")
	responseUploadsFirstPage := testutils.DataFromFile(t, "read/uploads-first-page.json")
	responseTeams := testutils.DataFromFile(t, "read/teams.json")
	responseBadRequest := testutils.DataFromFile(t, "read/error-bad-request.json")
	responseInvalidToken := testutils.DataFromFile(t, "read/error-invalid-token.json")
	responseNotFound := testutils.DataFromFile(t, "read/error-not-found.json")

	apiVersion := mockcond.Header(http.Header{"Public-Api-Version": []string{"2026-10-01"}})

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "recordings"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "stories", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Recordings are listed with the API version, without a page size, and link to the next cursor",
			Input: common.ReadParams{
				ObjectName: "recordings",
				Fields:     connectors.Fields("title", "start_datetime", "share_state"),
				PageSize:   10,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/_/public-api/v2/recordings"),
					apiVersion,
					mockcond.Body(`{}`),
				},
				Then: mockserver.Response(http.StatusOK, responseRecordingsFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "0470708a-df31-4386-9a66-f8b2ae480f70",
					Fields: map[string]any{
						"title":          "Ampersand proxy test",
						"start_datetime": "2026-10-01T14:04:29Z",
						"share_state":    "public",
					},
					Raw: map[string]any{
						"media_type":  "transcript",
						"sensitivity": "not_classified",
					},
				}},
				NextPage: "ApJNWoNoBHcCdjJ3DWNocm9ub2xvZ2ljYWx3BGRlc2NoAm4HAO5FaSm2QgZuBwAVPC0qtkIG",
				Done:     false,
			},
		},
		{
			Name: "Recordings include only the requested sections, send the cursor back and ignore Since",
			Input: common.ReadParams{
				ObjectName: "recordings",
				Fields:     connectors.Fields("title", "participants", "ai_template_sections"),
				NextPage:   "ApJNWoNoBHcCdjJ3DWNocm9ub2xvZ2ljYWx3BGRlc2NoAm4HAO5FaSm2QgZuBwAVPC0qtkIG",
				Since:      time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/_/public-api/v2/recordings"),
					mockcond.Body(`{
						"cursor": "ApJNWoNoBHcCdjJ3DWNocm9ub2xvZ2ljYWx3BGRlc2NoAm4HAO5FaSm2QgZuBwAVPC0qtkIG",
						"include": {"participants": true, "ai_template_sections": {"format": "json"}}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseRecordingsLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "0470708a-df31-4386-9a66-f8b2ae480f70",
					Fields: map[string]any{
						"title":                "Ampersand proxy test",
						"participants":         []any{},
						"ai_template_sections": []any{},
					},
					Raw: map[string]any{"ai_summary": nil},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Uploads keep those updated within Since and Until, both inclusive",
			Input: common.ReadParams{
				ObjectName: "uploads",
				Fields:     connectors.Fields("title", "status", "attributes"),
				Since:      time.Date(2026, 10, 28, 11, 37, 48, 0, time.UTC),
				Until:      time.Date(2026, 10, 28, 11, 37, 48, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/_/public-api/v2/uploads"),
					apiVersion,
					mockcond.Body(`{"include": {"attributes": true, "transcript": true}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseUploads),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "0efac057-7d26-4d82-afc7-771fc081aa51",
					Fields: map[string]any{
						"title":  "Ampersand schema validation (staged, updated)",
						"status": "staged",
						"attributes": map[string]any{
							"title":          "Ampersand schema validation (staged, updated)",
							"start_datetime": "2026-10-28T11:37:44.405272Z",
							"participants": []any{
								map[string]any{"name": "Speaker One", "speaker_key": "s1"},
							},
							"transcript": map[string]any{
								"language": "en",
								"segments": []any{
									map[string]any{
										"speaker_key": "s1",
										"start":       float64(0),
										"end":         float64(1200),
										"text":        "Schema validation.",
									},
								},
								"anonymous_speakers": []any{},
							},
							"team_id":      nil,
							"recorder_ids": []any{"e4bac6b2-8272-4db0-a4ec-892e3010631f"},
						},
					},
					Raw: map[string]any{"updated_datetime": "2026-10-28T11:37:48Z"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Uploads keep paging when every record on the page is filtered out",
			Input: common.ReadParams{
				ObjectName: "uploads",
				Fields:     connectors.Fields("title"),
				Since:      time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/_/public-api/v2/uploads"),
					mockcond.Body(`{}`),
				},
				Then: mockserver.Response(http.StatusOK, responseUploadsFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "ApJNWoNoBHcCdjJ3DWNocm9ub2xvZ2ljYWx3BGRlc2NoAm4HAO5FaSm2QgZuBwAVPC0qtkIG",
				Done:     false,
			},
		},
		{
			Name:  "Teams are not paginated",
			Input: common.ReadParams{ObjectName: "teams", Fields: connectors.Fields("name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/_/public-api/v2/teams"),
					apiVersion,
				},
				Then: mockserver.Response(http.StatusOK, responseTeams),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "9d228ffa-e5f7-4dd7-ab0a-5169fb878e21",
					Fields: map[string]any{"name": "Withampersand"},
					Raw:    map[string]any{"id": "9d228ffa-e5f7-4dd7-ab0a-5169fb878e21"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Validation error is parsed",
			Input: common.ReadParams{ObjectName: "recordings", Fields: connectors.Fields("id"), NextPage: "expired"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("bad_request: invalid_cursor"),
			},
		},
		{
			Name:  "Authentication error description is parsed",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseInvalidToken),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("invalid_token: The access token is invalid or expired."),
			},
		},
		{
			Name:  "Error code without details is parsed",
			Input: common.ReadParams{ObjectName: "hooks", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError("not_found"),
			},
		},
		{
			// Grain answers a body that is not valid JSON with a plain HTML page.
			Name:  "HTML error page is reported as a caller error",
			Input: common.ReadParams{ObjectName: "meeting_types", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.ResponseString(http.StatusBadRequest, "Bad Request"),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
