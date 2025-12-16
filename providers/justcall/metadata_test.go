package justcall

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	responseCustomFields := testutils.DataFromFile(t, "read/sales_dialer_contacts/custom-fields.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successful metadata for sales_dialer/contacts with custom fields",
			Input: []string{"sales_dialer/contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2.1/sales_dialer/contacts/custom-fields"),
				Then:  mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"sales_dialer/contacts": {
						DisplayName: "Sales Dialer Contacts",
						Fields: map[string]common.FieldMetadata{
							// Custom fields from the API response
							"membership_status": {
								DisplayName:  "membership_status",
								ValueType:    "string",
								ProviderType: "string",
							},
							"priority_level": {
								DisplayName:  "priority_level",
								ValueType:    "float",
								ProviderType: "number",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Objects without custom fields return metadata without custom fields",
			Input: []string{"users", "contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2.1/sales_dialer/contacts/custom-fields"),
				Then:  mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields:      map[string]common.FieldMetadata{},
					},
					"contacts": {
						DisplayName: "Contacts",
						Fields:      map[string]common.FieldMetadata{},
					},
				},
				Errors: map[string]error{},
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
