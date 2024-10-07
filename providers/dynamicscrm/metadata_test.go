package dynamicscrm

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
			Name:         "Mime response header expected",
			Input:        []string{"accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
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
			Server: mockserver.Reactive{
				Setup:     mockserver.ContentJSON(),
				Condition: mockcond.PathSuffix("EntityDefinitions(LogicalName='butterfly')"),
				OnSuccess: mockserver.Response(http.StatusOK, responseContactsSchema),
				OnFailure: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			ExpectedErrs: []error{ErrObjectNotFound},
		},
		{
			Name:  "Object doesn't have attributes",
			Input: []string{"accounts"},
			Server: mockserver.Crossroad{
				Setup: mockserver.ContentJSON(),
				Paths: []mockserver.Path{{
					Condition: mockcond.PathSuffix("EntityDefinitions(LogicalName='account')"),
					OnSuccess: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					Condition: mockcond.PathSuffix("EntityDefinitions(LogicalName='account')/Attributes"),
					OnSuccess: mockserver.ResponseString(http.StatusOK, `{"value":[]}`),
				}},
				OnFailure: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			ExpectedErrs: []error{ErrObjectMissingAttributes},
		},
		{
			Name:  "Correctly list metadata for account leads and invite contact",
			Input: []string{"contacts"},
			Server: mockserver.Crossroad{
				Setup: mockserver.ContentJSON(),
				Paths: []mockserver.Path{{
					Condition: mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')"),
					OnSuccess: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					Condition: mockcond.PathSuffix("EntityDefinitions(LogicalName='contact')/Attributes"),
					OnSuccess: mockserver.Response(http.StatusOK, responseContactsAttributes),
				}},
				OnFailure: mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
