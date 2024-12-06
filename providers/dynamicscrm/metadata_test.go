package dynamicscrm

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContactsSchema := testutils.DataFromFile(t, "contacts-schema.json")
	// Attributes file is a shorter form of real Microsoft server response.
	responseContactsAttributes := testutils.DataFromFile(t, "contacts-attributes.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Schema endpoint is not available for object",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsSchema),
			}.Server(),
			ExpectedErrs: []error{ErrObjectNotFound},
		},
		{
			Name:  "Attributes endpoint is not available for object",
			Input: []string{"butterflies"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("EntityDefinitions(LogicalName='butterfly')"),
				Then:  mockserver.Response(http.StatusOK, responseContactsSchema),
				Else:  mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			ExpectedErrs: []error{ErrObjectNotFound},
		},
		{
			Name:  "Object doesn't have attributes",
			Input: []string{"accounts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='account')"),
					Then: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='account')/Attributes"),
					Then: mockserver.ResponseString(http.StatusOK, `{"value":[]}`),
				}},
				Default: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			ExpectedErrs: []error{ErrObjectMissingAttributes},
		},
		{
			Name:  "Correctly list metadata for account leads and invite contact",
			Input: []string{"contacts"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')"),
					Then: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes"),
					Then: mockserver.Response(http.StatusOK, responseContactsAttributes),
				}},
				Default: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "contacts",
						FieldsMap: map[string]string{
							// nice display names
							"adx_publicprofilecopy":    "Public Profile Copy",
							"adx_identity_newpassword": "New Password Input",
							"department":               "Department",
							"shippingmethodcode":       "Shipping Method",
							"lastname":                 "Last Name",
							// schema name was used for display
							"leadsourcecodename": "LeadSourceCodeName",
							// underscore prefixed fields
							"_accountid_value": "Account",
							"_createdby_value": "Created By",
						},
					},
				},
				Errors: nil,
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
