package zoominfo

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const (
	objNews             = "news"
	objCompanyRankings  = "company-rankings"
	objAudiences        = "audiences"
	objCustomerSettings = "customer-settings"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")
	newsResponse := testutils.DataFromFile(t, "news.json")
	companyRankingsResponse := testutils.DataFromFile(t, "company-rankings.json")
	audiencesResponse := testutils.DataFromFile(t, "audiences.json")
	customerSettingsResponse := testutils.DataFromFile(t, "customer-settings.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Search (POST) and lookup (GET) objects sampled from data[]",
			Input: []string{objContacts, objNews, objCompanyRankings},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/gtm/data/v1/contacts/search"),
						mockcond.Body(`{"data":{"type":"ContactSearch","attributes":{"lastUpdatedDateAfter":"1970-01-01"}}}`),
					},
					Then: mockserver.Response(http.StatusOK, contactsResponse),
				}, {
					If: mockcond.And{
						mockcond.Path("/gtm/data/v1/news/search"),
						mockcond.Body(`{"data":{"type":"NewsSearch","attributes":{"pageDateMin":"1970-01-01"}}}`),
					},
					Then: mockserver.Response(http.StatusOK, newsResponse),
				}, {
					If:   mockcond.Path("/gtm/data/v1/lookup/company-rankings"),
					Then: mockserver.Response(http.StatusOK, companyRankingsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					objContacts: {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id":        {DisplayName: "id", ValueType: common.ValueTypeString},
							"type":      {DisplayName: "type", ValueType: common.ValueTypeString},
							"firstName": {DisplayName: "firstName", ValueType: common.ValueTypeString},
							"lastName":  {DisplayName: "lastName", ValueType: common.ValueTypeString},
							"jobTitle":  {DisplayName: "jobTitle", ValueType: common.ValueTypeString},
							"hasEmail":  {DisplayName: "hasEmail", ValueType: common.ValueTypeBoolean},
						},
					},
					objNews: {
						DisplayName: "News",
						Fields: map[string]common.FieldMetadata{
							"id":       {DisplayName: "id", ValueType: common.ValueTypeString},
							"type":     {DisplayName: "type", ValueType: common.ValueTypeString},
							"title":    {DisplayName: "title", ValueType: common.ValueTypeString},
							"url":      {DisplayName: "url", ValueType: common.ValueTypeString},
							"pageDate": {DisplayName: "pageDate", ValueType: common.ValueTypeString},
						},
					},
					objCompanyRankings: {
						DisplayName: "Company Rankings",
						Fields: map[string]common.FieldMetadata{
							"id":   {DisplayName: "id", ValueType: common.ValueTypeString},
							"type": {DisplayName: "type", ValueType: common.ValueTypeString},
							"name": {DisplayName: "name", ValueType: common.ValueTypeString},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "GET list object and singleton object sampled correctly",
			Input: []string{objAudiences, objCustomerSettings},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/gtm/studio/v1/audiences"),
					Then: mockserver.Response(http.StatusOK, audiencesResponse),
				}, {
					If:   mockcond.Path("/gtm/copilot/v1/customer-settings"),
					Then: mockserver.Response(http.StatusOK, customerSettingsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					objAudiences: {
						DisplayName: "Audiences",
						Fields: map[string]common.FieldMetadata{
							"id":       {DisplayName: "id", ValueType: common.ValueTypeString},
							"name":     {DisplayName: "name", ValueType: common.ValueTypeString},
							"rowCount": {DisplayName: "rowCount", ValueType: common.ValueTypeFloat},
						},
					},
					objCustomerSettings: {
						DisplayName: "Customer Settings",
						Fields: map[string]common.FieldMetadata{
							"id":                {DisplayName: "id", ValueType: common.ValueTypeString},
							"timezone":          {DisplayName: "timezone", ValueType: common.ValueTypeString},
							"enrichmentEnabled": {DisplayName: "enrichmentEnabled", ValueType: common.ValueTypeBoolean},
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
				return constructTestConnector(tt.Server)
			})
		})
	}
}

// TestListObjectMetadataUnknownObject verifies that an unsupported object does
// not abort the whole call but is recorded per-object in the Errors map.
func TestListObjectMetadataUnknownObject(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector(mockserver.Dummy())
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	result, err := conn.ListObjectMetadata(t.Context(), []string{"unicorns"})
	if err != nil {
		t.Fatalf("expected no top-level error, got %v", err)
	}

	objErr, ok := result.Errors["unicorns"]
	if !ok {
		t.Fatal("expected an entry in Errors for the unknown object")
	}

	if !errors.Is(objErr, common.ErrObjectNotSupported) {
		t.Fatalf("expected ErrObjectNotSupported, got %v", objErr)
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	// Redirect calls to the mock server.
	connector.SetBaseURL(server.URL)

	return connector, nil
}
