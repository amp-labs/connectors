package klaviyo

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorUnsupportedPagination := testutils.DataFromFile(t, "read-unsupported-pagination.json")
	responseCampaigns := testutils.DataFromFile(t, "read-campaigns.json")
	responseProfilesFirstPage := testutils.DataFromFile(t, "read-profiles-1-first-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "orders", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Error response when pagination is not available for the object",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentMIME("application/vnd.api+json"),
				Always: mockserver.Response(http.StatusBadRequest, errorUnsupportedPagination),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Invalid input: 'page_size' is not a valid field for the resource 'list'."),
			},
		},
		{
			Name: "Profiles first page has a link to the next",
			Input: common.ReadParams{
				ObjectName: "profiles",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/vnd.api+json"),
				If:    mockcond.PathSuffix("/api/profiles"),
				Then:  mockserver.Response(http.StatusOK, responseProfilesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email": "jennifer@gmail.com",
					},
					Raw: map[string]any{
						"id": "01HSXWNWF52J5PJG45BW383RMV",
						"attributes": map[string]any{
							"email":           "jennifer@gmail.com",
							"phone_number":    nil,
							"external_id":     nil,
							"anonymous_id":    nil,
							"first_name":      nil,
							"last_name":       nil,
							"organization":    nil,
							"locale":          nil,
							"title":           nil,
							"image":           nil,
							"created":         "2024-03-26T17:24:42+00:00",
							"updated":         "2024-03-26T17:24:42+00:00",
							"last_event_date": "2024-03-26T17:24:41+00:00",
							"location":        map[string]any{},
							"properties":      map[string]any{},
						},
					},
				}},
				NextPage: "https://a.klaviyo.com/api/profiles?page%5Bsize%5D=1&page%5Bcursor%5D=bmV4dDo6aWQ6OjAxSFNYV05XRjUySjVQSkc0NUJXMzgzUk1W", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of campaigns with required filter",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2024, 3, 4, 8, 22, 56, 0, time.UTC),
				Filter:     "equals(messages.channel,'email')",
			},
			Comparator: testroutines.ComparatorSubsetRead,
			Server: mockserver.Conditional{
				Setup: mockserver.ContentMIME("application/vnd.api+json"),
				If: mockcond.And{
					mockcond.PathSuffix("/api/campaigns"),
					mockcond.QueryParam("filter",
						"greater-than(updated_at,2024-03-04T08:22:56Z),equals(messages.channel,'email')"),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaigns),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Email Campaign - Nov 15, 2024, 1:18 AM",
					},
					Raw: map[string]any{
						"attributes": map[string]any{
							"name":             "Email Campaign - Nov 15, 2024, 1:18 AM",
							"status":           "Scheduled",
							"archived":         false,
							"audiences":        map[string]any{},
							"send_options":     map[string]any{},
							"tracking_options": map[string]any{},
							"send_strategy":    map[string]any{},
							"created_at":       "2024-11-14T23:18:34.827140+00:00",
							"scheduled_at":     "2024-11-14T23:20:02.718919+00:00",
							"updated_at":       "2024-11-14T23:20:32.232276+00:00",
							"send_time":        "2024-11-30T18:15:00+00:00",
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetURL(serverURL)

	return connector, nil
}
