package auphan

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	responsePaymentsType := testutils.DataFromFile(t, "types-get-payments.json")
	responseStores := testutils.DataFromFile(t, "stores-page-1.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Report schema comes from reports/types.get, stores from a sampled record",
			Input: []string{"payments", "stores"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/v1.php/reports/types.get"),
						mockcond.Body(`{"store_id": 73921, "params": {"type": "payments"}}`),
					},
					Then: mockserver.Response(http.StatusOK, responsePaymentsType),
				}, {
					If:   mockcond.Path("/api/v1.php/settings/stores.list"),
					Then: mockserver.Response(http.StatusOK, responseStores),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"payments": {
						DisplayName: "Payments",
						Fields: map[string]common.FieldMetadata{
							"payment_type": {
								DisplayName:  "Payment Type",
								ValueType:    common.ValueTypeString,
								ProviderType: "dimension",
							},
							"payment": {
								DisplayName:  "Payment",
								ValueType:    common.ValueTypeString,
								ProviderType: "measure",
							},
						},
					},
					"stores": {
						DisplayName: "Stores",
						Fields: map[string]common.FieldMetadata{
							"store_id": {DisplayName: "store_id", ValueType: common.ValueTypeOther},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "success false with HTTP 200 fails the object",
			Input: []string{"sales"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "error-validation.json")),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{"sales": common.ErrBadRequest},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
