package fourfour

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

	contactsResponse := testutils.DataFromFile(t, "contacts.json")
	leadsResponse := testutils.DataFromFile(t, "leads.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"contacts", "leads"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/odata/contacts"),
					Then: mockserver.Response(http.StatusOK, contactsResponse),
				}, {
					If:   mockcond.Path("/odata/leads"),
					Then: mockserver.Response(http.StatusOK, leadsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id":          {DisplayName: "id", ValueType: common.ValueTypeString},
							"first_name":  {DisplayName: "first_name", ValueType: common.ValueTypeString},
							"last_name":   {DisplayName: "last_name", ValueType: common.ValueTypeString},
							"email":       {DisplayName: "email", ValueType: common.ValueTypeString},
							"phone":       {DisplayName: "phone", ValueType: common.ValueTypeString},
							"title":       {DisplayName: "title", ValueType: common.ValueTypeString},
							"account_id":  {DisplayName: "account_id", ValueType: common.ValueTypeString},
							"department":  {DisplayName: "department", ValueType: common.ValueTypeString},
							"lead_source": {DisplayName: "lead_source", ValueType: common.ValueTypeString},
							"owner_id":    {DisplayName: "owner_id", ValueType: common.ValueTypeString},
							"region":      {DisplayName: "region", ValueType: common.ValueTypeString},
							"created":     {DisplayName: "created", ValueType: common.ValueTypeString},
							"updated":     {DisplayName: "updated", ValueType: common.ValueTypeString},
						},
					},
					"leads": {
						DisplayName: "Leads",
						Fields: map[string]common.FieldMetadata{
							"id":                       {DisplayName: "id", ValueType: common.ValueTypeString},
							"first_name":               {DisplayName: "first_name", ValueType: common.ValueTypeString},
							"last_name":                {DisplayName: "last_name", ValueType: common.ValueTypeString},
							"title":                    {DisplayName: "title", ValueType: common.ValueTypeString},
							"company":                  {DisplayName: "company", ValueType: common.ValueTypeString},
							"email":                    {DisplayName: "email", ValueType: common.ValueTypeString},
							"phone":                    {DisplayName: "phone", ValueType: common.ValueTypeString},
							"website":                  {DisplayName: "website", ValueType: common.ValueTypeString},
							"lead_source":              {DisplayName: "lead_source", ValueType: common.ValueTypeString},
							"status":                   {DisplayName: "status", ValueType: common.ValueTypeString},
							"industry":                 {DisplayName: "industry", ValueType: common.ValueTypeString},
							"annual_revenue":           {DisplayName: "annual_revenue", ValueType: common.ValueTypeFloat},
							"number_of_employees":      {DisplayName: "number_of_employees", ValueType: common.ValueTypeFloat},
							"owner_id":                 {DisplayName: "owner_id", ValueType: common.ValueTypeString},
							"is_converted":             {DisplayName: "is_converted", ValueType: common.ValueTypeBoolean},
							"converted_date":           {DisplayName: "converted_date", ValueType: common.ValueTypeString},
							"converted_account_id":     {DisplayName: "converted_account_id", ValueType: common.ValueTypeString},
							"converted_contact_id":     {DisplayName: "converted_contact_id", ValueType: common.ValueTypeString},
							"converted_opportunity_id": {DisplayName: "converted_opportunity_id", ValueType: common.ValueTypeString},
							"account_id":               {DisplayName: "account_id", ValueType: common.ValueTypeString},
							"region":                   {DisplayName: "region", ValueType: common.ValueTypeString},
							"created":                  {DisplayName: "created", ValueType: common.ValueTypeString},
							"updated":                  {DisplayName: "updated", ValueType: common.ValueTypeString},
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
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(serverURL)

	return connector, nil
}
