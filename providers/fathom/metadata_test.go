package fathom

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	meetingsResponse := testutils.DataFromFile(t, "meetings-first-page.json")
	teamsResponse := testutils.DataFromFile(t, "teams.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"meetings", "teams"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/external/v1/meetings"),
					Then: mockserver.Response(http.StatusOK, meetingsResponse),
				}, {
					If:   mockcond.Path("/external/v1/teams"),
					Then: mockserver.Response(http.StatusOK, teamsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"meetings": {
						DisplayName: "Meetings",
						Fields:      nil,
						FieldsMap: map[string]string{
							"action_items":         "action_items",
							"calendar_invitees":    "calendar_invitees",
							"created_at":           "created_at",
							"crm_matches":          "crm_matches",
							"default_summary":      "default_summary",
							"meeting_title":        "meeting_title",
							"meeting_type":         "meeting_type",
							"recorded_by":          "recorded_by",
							"recording_end_time":   "recording_end_time",
							"recording_start_time": "recording_start_time",
							"scheduled_end_time":   "scheduled_end_time",
							"scheduled_start_time": "scheduled_start_time",
							"share_url":            "share_url",
							"title":                "title",
							"transcript":           "transcript",
							"transcript_language":  "transcript_language",
							"url":                  "url",
						},
					},
					"teams": {
						DisplayName: "Teams",
						Fields:      nil,
						FieldsMap: map[string]string{
							"created_at": "created_at",
							"name":       "name",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
