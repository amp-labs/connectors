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

// objContacts is defined in objects.go (package-level).
const (
	objEnrichContacts  = "enrich-contacts"
	objCompanyRankings = "company-rankings"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	contactFieldsResponse := testutils.DataFromFile(t, "lookup-search-contact-output.json")
	enrichContactFieldsResponse := testutils.DataFromFile(t, "lookup-enrich-contact-output.json")
	companyRankingsResponse := testutils.DataFromFile(t, "company-rankings.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Search and enrich objects via lookup output fields; lookup object sampled",
			Input: []string{objContacts, objEnrichContacts, objCompanyRankings},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/gtm/data/v1/lookup/search"),
						mockcond.QueryParam("filter[entity]", "contact"),
						mockcond.QueryParam("filter[fieldType]", "output"),
					},
					Then: mockserver.Response(http.StatusOK, contactFieldsResponse),
				}, {
					If: mockcond.And{
						mockcond.Path("/gtm/data/v1/lookup/enrich"),
						mockcond.QueryParam("filter[entity]", "contact"),
						mockcond.QueryParam("filter[fieldType]", "output"),
					},
					Then: mockserver.Response(http.StatusOK, enrichContactFieldsResponse),
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
							"firstName":       {DisplayName: "firstName", ValueType: common.ValueTypeOther},
							"lastName":        {DisplayName: "lastName", ValueType: common.ValueTypeOther},
							"companyId":       {DisplayName: "companyId", ValueType: common.ValueTypeOther},
							"managementLevel": {DisplayName: "managementLevel", ValueType: common.ValueTypeOther},
						},
					},
					objEnrichContacts: {
						DisplayName: "Enrich Contacts",
						Fields: map[string]common.FieldMetadata{
							"id":    {DisplayName: "id", ValueType: common.ValueTypeOther},
							"email": {DisplayName: "email", ValueType: common.ValueTypeOther},
							"phone": {DisplayName: "phone", ValueType: common.ValueTypeOther},
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
