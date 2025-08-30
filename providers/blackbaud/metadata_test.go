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
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"id":                    "id",
							"name":                  "name",
							"iso_4217":              "iso_4217",
							"locale":                "locale",
							"decimal_digits":        "decimal_digits",
							"currency_symbol":       "currency_symbol",
							"rounding_type":         "rounding_type",
							"active":                "active",
							"organization_currency": "organization_currency",
						},
					},
					"crm-volmg/volunteers": {
						DisplayName: "Crm-Volmg/Volunteers",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"id":                    "id",
							"name":                  "name",
							"address":               "address",
							"city":                  "city",
							"state":                 "state",
							"post_code":             "post_code",
							"lookup_id":             "lookup_id",
							"constituent_type":      "constituent_type",
							"sort_constituent_name": "sort_constituent_name",
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
		Metadata: map[string]string{
			"Bb-Api-Subscription-Key": "d747f7eca52d495998eef6e4bc923147",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server.
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
