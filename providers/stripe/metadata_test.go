package stripe

import (
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"coupons", "products"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"coupons": {
						DisplayName: "Coupons",
						FieldsMap: map[string]string{
							"id":               "Id",
							"livemode":         "Livemode",
							"currency":         "Currency",
							"currency_options": "Currency Options",
							"percent_off":      "Percent Off",
						},
					},
					"products": {
						DisplayName: "Products",
						FieldsMap: map[string]string{
							"id":            "Id",
							"images":        "Images",
							"default_price": "Default Price",
							"tax_code":      "Tax Code",
							"unit_label":    "Unit Label",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:       "Checkout Sessions",
			Input:      []string{"checkout/sessions"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"checkout/sessions": {
						DisplayName: "Payment Checkout Sessions",
						Fields: map[string]common.FieldMetadata{
							"line_items": {
								DisplayName:  "Line Items",
								ValueType:    "other",
								ProviderType: "object",
							},
							"currency": {
								DisplayName:  "Currency",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: nil,
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: server.Client(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(server.URL)

	return connector, nil
}
