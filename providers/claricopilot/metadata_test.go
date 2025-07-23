package claricopilot

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

	usersResponse := testutils.DataFromFile(t, "users.json")
	callsResponse := testutils.DataFromFile(t, "calls.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"users", "calls"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/users"),
					Then: mockserver.Response(http.StatusOK, usersResponse),
				}, {
					If:   mockcond.Path("/calls"),
					Then: mockserver.Response(http.StatusOK, callsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"calls": {
						DisplayName: "Calls",
						Fields:      nil,
						FieldsMap: map[string]string{
							"audio_url":            "audio_url",
							"bot_not_join_reason":  "bot_not_join_reason",
							"call_review_page_url": "call_review_page_url",
							"disposition":          "disposition",
							"externalParticipants": "externalParticipants",
							"id":                   "id",
							"joinedParticipants":   "joinedParticipants",
							"last_modified_time":   "last_modified_time",
							"metrics":              "metrics",
							"source_id":            "source_id",
							"status":               "status",
							"time":                 "time",
							"title":                "title",
							"type":                 "type",
							"users":                "users",
							"video_url":            "video_url",
						},
					},
					"users": {
						DisplayName: "Users",
						Fields:      nil,
						FieldsMap: map[string]string{
							"email":        "email",
							"id":           "id",
							"is_recording": "is_recording",
							"name":         "name",
							"role":         "role",
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
