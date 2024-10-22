package instantly

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseInvalidPath := testutils.DataFromFile(t, "invalid-path.json")
	responseCampaigns := testutils.DataFromFile(t, "read-campaigns.json")
	responseTags := testutils.DataFromFile(t, "read-tags.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "campaigns"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.ReadParams{ObjectName: "campaigns", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
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
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "campaigns", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseInvalidPath),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Not Found"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "emails", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"yourEmails": []
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "emails", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"data": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name: "Next page is correctly calculated",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				NextPage:   "test-placeholder?skip=700",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.NewServer(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Create fake response big enough to conclude that next page exists.
				manyCampaigns := make([]string, DefaultPageSize)
				for i := 0; i < DefaultPageSize; i++ {
					manyCampaigns[i] = "{}"
				}
				data := fmt.Sprintf("[%v]", strings.Join(manyCampaigns, ","))
				_, _ = w.Write([]byte(data))
			}),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				NextPage: "test-placeholder?limit=100&skip=800",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Current empty page signifies no next page",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, "[]"),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)

				return actual.NextPage.String() == expectedNextPage &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				NextPage: "{{testServerURL}}/v1/campaign/list?limit=100&skip=100",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read campaigns with chosen fields",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseCampaigns),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Second Campaign",
					},
					Raw: map[string]any{
						"id":   "27dd47ff-5a78-4377-a1eb-98f593f37219",
						"name": "Second Campaign",
					},
				}, {
					Fields: map[string]any{
						"name": "My Campaign",
					},
					Raw: map[string]any{
						"id":   "65d890fe-ae4d-43d0-b014-af0e89b6281f",
						"name": "My Campaign",
					},
				}},
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read tags with chosen fields",
			Input: common.ReadParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("label", "description"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"label":       "High Delivery 3",
						"description": "High Delivery Accounts 3",
					},
					Raw: map[string]any{
						"organization_id": "803a064c-a636-49fe-bc45-5043da7a4ee7",
						"label":           "High Delivery 3",
						"description":     "High Delivery Accounts 3",
					},
				}, {
					Fields: map[string]any{
						"label":       "High Delivery 2",
						"description": "High Delivery Accounts 2",
					},
					Raw: map[string]any{
						"organization_id": "803a064c-a636-49fe-bc45-5043da7a4ee7",
						"label":           "High Delivery 2",
						"description":     "High Delivery Accounts 2",
					},
				}},
				Done: false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
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
	connector.setBaseURL(serverURL)

	return connector, nil
}
