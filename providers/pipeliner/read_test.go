package pipeliner

import (
	"errors"
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

	responseNotFound := testutils.DataFromFile(t, "resource-not-found.json")
	responseProfilesFirstPage := testutils.DataFromFile(t, "read-profiles-1-first-page.json")
	responseProfilesSecondPage := testutils.DataFromFile(t, "read-profiles-2-second-page.json")
	responseProfilesLastPage := testutils.DataFromFile(t, "read-profiles-3-last-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "Profiles"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("The requested URL was not found on the server"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"garbage": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"data": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Next page cursor may be missing in payload",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "success": true,
				  "total": 0,
				  "data": []
				}`),
			}.Server(),
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is inferred, when provided with an object",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseProfilesFirstPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)

				return actual.NextPage.String() == expectedNextPage
			},
			Expected: &common.ReadResult{
				NextPage: "{{testServerURL}}/api/v100/rest/spaces/test-workspace/entities/Profiles?after=WyIwMDAwMDAwMC0wMDAwLTAwMDEtMDAwMS0wMDAwMDAwMDhlOTciXQ%3D%3D&first=100", //nolint:lll
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is empty, when provided with null object",
			Input: common.ReadParams{ObjectName: "Profiles", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseProfilesLastPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
			Expected:     &common.ReadResult{NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "Profiles",
				Fields:     connectors.Fields("name", "owner_id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseProfilesSecondPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				// custom comparison focuses on subset of fields to keep the test short
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":     "Lang_DefaultProfileAllUsers",
						"owner_id": "00000000-0000-0000-0000-000000008e97",
					},
					Raw: map[string]any{
						"name":       "Lang_DefaultProfileAllUsers",
						"owner_id":   "00000000-0000-0000-0000-000000008e97",
						"use_lang":   true,
						"entity":     float64(3),
						"is_deleted": false,
					},
				}, {
					Fields: map[string]any{
						"name":     "Lang_DefaultProfileMy",
						"owner_id": "00000000-0000-0000-0000-000000008e97",
					},
					Raw: map[string]any{
						"name":       "Lang_DefaultProfileMy",
						"owner_id":   "00000000-0000-0000-0000-000000008e97",
						"use_lang":   true,
						"entity":     float64(3),
						"is_deleted": false,
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
		WithWorkspace("test-workspace"),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.WithBaseURL(serverURL)

	return connector, nil
}
