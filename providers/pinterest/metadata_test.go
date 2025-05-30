package pinterest

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

	pinsResponse := testutils.DataFromFile(t, "pins.json")
	boardsResponse := testutils.DataFromFile(t, "boards.json")
	mediaResponse := testutils.DataFromFile(t, "media.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"pins", "boards", "media"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v5/pins"),
					Then: mockserver.Response(http.StatusOK, pinsResponse),
				}, {
					If:   mockcond.Path("/v5/boards"),
					Then: mockserver.Response(http.StatusOK, boardsResponse),
				}, {
					If:   mockcond.Path("/v5/media"),
					Then: mockserver.Response(http.StatusOK, mediaResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"pins": {
						DisplayName: "Pins",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":                "id",
							"created_at":        "created_at",
							"link":              "link",
							"title":             "title",
							"description":       "description",
							"dominant_color":    "dominant_color",
							"alt_text":          "alt_text",
							"creative_type":     "creative_type",
							"board_id":          "board_id",
							"board_section_id":  "board_section_id",
							"board_owner":       "board_owner",
							"is_owner":          "is_owner",
							"media":             "media",
							"parent_pin_id":     "parent_pin_id",
							"is_standard":       "is_standard",
							"has_been_promoted": "has_been_promoted",
							"note":              "note",
							"pin_metrics":       "pin_metrics",
						},
					},
					"boards": {
						DisplayName: "Boards",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":                     "id",
							"created_at":             "created_at",
							"board_pins_modified_at": "board_pins_modified_at",
							"name":                   "name",
							"description":            "description",
							"collaborator_count":     "collaborator_count",
							"pin_count":              "pin_count",
							"follower_count":         "follower_count",
							"media":                  "media",
							"owner":                  "owner",
							"privacy":                "privacy",
						},
					},
					"media": {
						DisplayName: "Media",
						Fields:      nil,
						FieldsMap: map[string]string{
							"media_id":   "media_id",
							"media_type": "media_type",
							"status":     "status",
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
