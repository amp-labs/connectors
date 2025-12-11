package zoho

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	dealResponse := testutils.DataFromFile(t, "deals.json")
	unsupported := testutils.DataFromFile(t, "arsenal.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Unsupported objects",
			Input: []string{"arsenal"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("module", "Arsenal"),
					Then: mockserver.Response(http.StatusBadRequest, unsupported),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"arsenal": mockutils.ExpectedSubsetErrors{
						common.ErrCaller,
						errors.New(string(unsupported)),
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Supported objects",
			Input: []string{"Deals"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("module", "Deals"),
					Then: mockserver.Response(http.StatusOK, dealResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Deals": {
						DisplayName: "Deals",
						Fields: map[string]common.FieldMetadata{
							"Owner": {
								DisplayName:  "Deal Owner",
								ValueType:    "other",
								ProviderType: "ownerlookup",
								ReadOnly:     goutils.Pointer(true),
								Values:       nil,
							},
							"Stage": {
								DisplayName:  "Stage",
								ValueType:    "singleSelect",
								ProviderType: "picklist",
								ReadOnly:     goutils.Pointer(false),
								Values: common.FieldValues{
									{
										Value:        "Qualification",
										DisplayValue: "Qualification",
									}, {
										Value:        "Needs Analysis",
										DisplayValue: "Needs Analysis",
									}, {
										Value:        "Value Proposition",
										DisplayValue: "Value Proposition",
									}, {
										Value:        "Id. Decision Makers",
										DisplayValue: "Identify Decision Makers",
									}, {
										Value:        "Proposal/Price Quote",
										DisplayValue: "Proposal/Price Quote",
									}, {
										Value:        "Negotiation/Review",
										DisplayValue: "Negotiation/Review",
									}, {
										Value:        "Closed Won",
										DisplayValue: "Closed Won",
									}, {
										Value:        "Closed Lost",
										DisplayValue: "Closed Lost",
									}, {
										Value:        "Closed Lost to Competition",
										DisplayValue: "Closed Lost to Competition",
									},
								},
							},
						},
					},
				},
				Errors: map[string]error{},
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
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
