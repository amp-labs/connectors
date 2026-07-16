package campaignmonitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	clientsResponse := testutils.DataFromFile(t, "clients.json")
	adminsResponse := testutils.DataFromFile(t, "admins.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"clients", "admins"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v3.3/clients.json"),
					Then: mockserver.Response(http.StatusOK, clientsResponse),
				}, {
					If:   mockcond.Path("/api/v3.3/admins.json"),
					Then: mockserver.Response(http.StatusOK, adminsResponse),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"clients": {
						DisplayName: "Clients",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"ClientID": "ClientID",
							"Name":     "Name",
						},
					},
					"admins": {
						DisplayName: "Admins",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"EmailAddress": "EmailAddress",
							"Name":         "Name",
							"Status":       "Status",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(server.URL)

	return connector, nil
}
