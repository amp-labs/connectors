package breakcold

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	statusResponse := testutils.DataFromFile(t, "status.json")
	leadsResponse := testutils.DataFromFile(t, "leads.json")

	tests := []testroutines.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"status", "leads"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/status"),
					Then: mockserver.Response(http.StatusOK, statusResponse),
				}, {
					If: mockcond.Path("/leads/list"),

					Then: mockserver.Response(http.StatusOK, leadsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"status": {
						DisplayName: "Status",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"type":         "type",
							"id":           "id",
							"name":         "name",
							"order":        "order",
							"color":        "color",
							"success_rate": "success_rate",
							"id_space":     "id_space",
						},
					},
					"leads": {
						DisplayName: "Leads",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"id":                   "id",
							"email":                "email",
							"company":              "company",
							"phone":                "phone",
							"linkedin_url":         "linkedin_url",
							"linkedin_company_url": "linkedin_company_url",
							"facebook_username":    "facebook_username",
							"youtube_username":     "youtube_username",
							"instagram_username":   "instagram_username",
							"telegram_username":    "telegram_username",
							"first_name":           "first_name",
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

			tt.Run(t, func() (testroutines.TestableMetadataReader, error) {
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

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(server.URL)

	return connector, nil
}
