package constantcontact

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseContactIDsError := testutils.DataFromFile(t, "read-contact-ids-error.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "read-contacts-1-first-page.json")
	responseContactsLastPage := testutils.DataFromFile(t, "read-contacts-2-second-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "activities"},
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
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "contact_id_xrefs", Fields: connectors.Fields("contact_id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseContactIDsError),
			}.Server(),
			ExpectedErrs: []error{
				errors.New("Sequence ID is invalid."), // nolint:goerr113
				common.ErrBadRequest,
			},
		},
		{
			Name: "Read contacts first page",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("first_name", "last_name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsFirstPage),
			}.Server(),
			Comparator: readComparator,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"first_name": "Debora",
						"last_name":  "Lang",
					},
					Raw: map[string]any{
						"company_name": "Acme Corp.",
						"contact_id":   "af73e650-96f0-11ef-b2a0-fa163eafb85e",
					},
				}},
				NextPage: "{{testServerURL}}/v3/contacts?cursor=bGltaXQ9MSZuZXh0PTI=",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("first_name", "last_name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsLastPage),
			}.Server(),
			Comparator: readComparator,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"first_name": "John",
						"last_name":  "Doe",
					},
					Raw: map[string]any{
						"create_source": "Account",
						"contact_id":    "832444c0-4392-11ef-95d3-fa163e761ca9",
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

func readComparator(baseURL string, actual, expected *common.ReadResult) bool {
	expectedNextPage := strings.ReplaceAll(expected.NextPage.String(), "{{testServerURL}}", baseURL)

	return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
		mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
		actual.NextPage.String() == expectedNextPage &&
		actual.Rows == expected.Rows &&
		actual.Done == expected.Done
}
