package jump

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "metadata-contacts.json")
	requestContacts := testutils.DataFromFile(t, "metadata/request/contacts.json")
	emptyContactsResponse := testutils.DataFromFile(t, "metadata-empty-contacts.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Sample contacts metadata from first list record",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Body(string(requestContacts)),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"contactInfo": {
								DisplayName: "contactInfo",
								ValueType:   common.ValueTypeOther,
							},
							"email": {
								DisplayName: "email",
								ValueType:   common.ValueTypeString,
							},
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"insertedAt": {
								DisplayName: "insertedAt",
								ValueType:   common.ValueTypeString,
							},
							"integrationReferences": {
								DisplayName: "integrationReferences",
								ValueType:   common.ValueTypeOther,
							},
							"name": {
								DisplayName: "name",
								ValueType:   common.ValueTypeString,
							},
							"status": {
								DisplayName: "status",
								ValueType:   common.ValueTypeString,
							},
							"type": {
								DisplayName: "type",
								ValueType:   common.ValueTypeString,
							},
						},
					},
				},
			},
		},
		{
			Name:  "Returns error when contacts list is empty",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Body(string(requestContacts)),
				},
				Then: mockserver.Response(http.StatusOK, emptyContactsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"contacts": common.ErrMissingExpectedValues,
				},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
