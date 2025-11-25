package blackbaud

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	currenciesResponse := testutils.DataFromFile(t, "currencies.json")
	volunteersResponse := testutils.DataFromFile(t, "volunteers.json")

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"crm-adnmg/currencies", "crm-volmg/volunteers"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/crm-adnmg/currencies/list"),
					Then: mockserver.Response(http.StatusOK, currenciesResponse),
				}, {
					If:   mockcond.Path("/crm-volmg/volunteers/search"),
					Then: mockserver.Response(http.StatusOK, volunteersResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"crm-adnmg/currencies": {
						DisplayName: "Crm-Adnmg/Currencies",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"iso_4217": {
								DisplayName: "iso_4217",
								ValueType:   "other",
							},
							"locale": {
								DisplayName: "locale",
								ValueType:   "other",
							},
							"decimal_digits": {
								DisplayName: "decimal_digits",
								ValueType:   "other",
							},
							"currency_symbol": {
								DisplayName: "currency_symbol",
								ValueType:   "other",
							},
							"rounding_type": {
								DisplayName: "rounding_type",
								ValueType:   "other",
							},
							"active": {
								DisplayName: "active",
								ValueType:   "other",
							},
							"organization_currency": {
								DisplayName: "organization_currency",
								ValueType:   "other",
							},
						},
						FieldsMap: map[string]string{},
					},
					"crm-volmg/volunteers": {
						DisplayName: "Crm-Volmg/Volunteers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"address": {
								DisplayName: "address",
								ValueType:   "other",
							},
							"city": {
								DisplayName: "city",
								ValueType:   "other",
							},
							"state": {
								DisplayName: "state",
								ValueType:   "other",
							},
							"post_code": {
								DisplayName: "post_code",
								ValueType:   "other",
							},
							"lookup_id": {
								DisplayName: "lookup_id",
								ValueType:   "other",
							},
							"constituent_type": {
								DisplayName: "constituent_type",
								ValueType:   "other",
							},
							"sort_constituent_name": {
								DisplayName: "sort_constituent_name",
								ValueType:   "other",
							},
						},
						FieldsMap: map[string]string{},
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

func TestConnectorStringMethods(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://mockserver.test")
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	testutils.CheckOutput(t, "conn.Provider():", "blackbaud", conn.Provider())
	testutils.CheckOutput(t, "conn.String():", "blackbaud.Connector[root]", conn.String())
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"bbApiSubscriptionKey": "d747f7eca52d495998eef6e4bc923147",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server.
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
