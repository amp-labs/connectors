package loxo

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	companiesResponse := testutils.DataFromFile(t, "companies.json")
	currenciesResponse := testutils.DataFromFile(t, "currencies.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"companies", "currencies"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/integration-user-loxo-withampersand-com/companies"),
					Then: mockserver.Response(http.StatusOK, companiesResponse),
				}, {
					If:   mockcond.Path("/integration-user-loxo-withampersand-com/currencies"),
					Then: mockserver.Response(http.StatusOK, currenciesResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"companies": {
						DisplayName: "Companies",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"url": {
								DisplayName: "url",
								ValueType:   "other",
							},
							"job_count": {
								DisplayName: "job_count",
								ValueType:   "other",
							},
							"owned_by_name": {
								DisplayName: "owned_by_name",
								ValueType:   "other",
							},
							"owned_by_id": {
								DisplayName: "owned_by_id",
								ValueType:   "other",
							},
						},
						FieldsMap: nil,
					},
					"currencies": {
						DisplayName: "Currencies",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"code": {
								DisplayName: "code",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"symbol": {
								DisplayName: "symbol",
								ValueType:   "other",
							},
							"precision": {
								DisplayName: "precision",
								ValueType:   "other",
							},
							"default": {
								DisplayName: "default",
								ValueType:   "other",
							},
							"position": {
								DisplayName: "position",
								ValueType:   "other",
							},
						},
						FieldsMap: nil,
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
		Workspace:           "pod4.app.loxo.co",
		Metadata: map[string]string{
			"agencySlug": "integration-user-loxo-withampersand-com",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(serverURL)

	return connector, nil
}
