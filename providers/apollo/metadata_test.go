package apollo

import (
	"errors"
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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")
	opportunityResponse := testutils.DataFromFile(t, "opportunities.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Product name instead of API documented object name",
			Input: []string{"deals"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, opportunityResponse),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"deals": {
						DisplayName: "deals",
						FieldsMap: map[string]string{
							"id":                  "id",
							"owner_id":            "owner_id",
							"team_id":             "team_id",
							"amount":              "amount",
							"salesforce_owner_id": "salesforce_owner_id",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"contacts", "opportunities", "arsenal"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/opportunities/search"),
					Then: mockserver.Response(http.StatusOK, opportunityResponse),
				}, {
					If:   mockcond.PathSuffix("/arsenal"),
					Then: mockserver.Response(http.StatusBadRequest, unsupportedResponse),
				}, {
					If:   mockcond.PathSuffix("/contacts/search"),
					Then: mockserver.Response(http.StatusOK, contactsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "contacts",
						FieldsMap: map[string]string{
							"id":                 "id",
							"account":            "account",
							"first_name":         "first_name",
							"last_name":          "last_name",
							"name":               "name",
							"city":               "city",
							"account_phone_note": "account_phone_note",
						},
					},
					"opportunities": {
						DisplayName: "opportunities",
						FieldsMap: map[string]string{
							"id":                  "id",
							"owner_id":            "owner_id",
							"team_id":             "team_id",
							"amount":              "amount",
							"salesforce_owner_id": "salesforce_owner_id",
						},
					},
				},
				Errors: map[string]error{
					"arsenal": mockutils.ExpectedSubsetErrors{
						common.ErrCaller,
						errors.New(string(unsupportedResponse)), // nolint:goerr113
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
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
