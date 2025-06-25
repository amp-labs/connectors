package campaignmonitor

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	clientsResponse := testutils.DataFromFile(t, "clients.json")
	adminsResponse := testutils.DataFromFile(t, "admins.json")
	campaignsResponse := testutils.DataFromFile(t, "campaigns.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"clients", "admins", "campaigns"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("api/v3.3/clients.json"),
					Then: mockserver.Response(http.StatusOK, clientsResponse),
				}, {
					If:   mockcond.Path("api/v3.3/admins.json"),
					Then: mockserver.Response(http.StatusOK, adminsResponse),
				}, {
					If:   mockcond.Path("api/v3.3/clients/744cdce058fc61d9ef5e2492f8d8fbaf/campaigns.json"),
					Then: mockserver.Response(http.StatusOK, campaignsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
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
					"campaigns": {
						DisplayName: "Campaigns",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"Name":              "Name",
							"FromName":          "FromName",
							"FromEmail":         "FromEmail",
							"ReplyTo":           "ReplyTo",
							"SentDate":          "SentDate",
							"TotalRecipients":   "TotalRecipients",
							"CampaignID":        "CampaignID",
							"Subject":           "Subject",
							"Tags":              "Tags",
							"WebVersionURL":     "WebVersionURL",
							"WebVersionTextURL": "WebVersionTextURL",
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
		AuthenticatedClient: http.DefaultClient,
		Metadata: map[string]string{
			"clientId": "744cdce058fc61d9ef5e2492f8d8fbaf",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(serverURL)

	return connector, nil
}
