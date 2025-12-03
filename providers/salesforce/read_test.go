package salesforce

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUnknownObject := testutils.DataFromFile(t, "unknown-object.json")
	responseLeadsFirstPage := testutils.DataFromFile(t, "read-list-leads.json")
	responseListContacts := testutils.DataFromFile(t, "read-list-contacts.json")
	responseOpportunityWithAccount := testutils.DataFromFile(t, "read-opportunity-with-account.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("Name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseUnknownObject),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("sObject type 'Accout' is not supported"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("Name")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"garbage": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("Name")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"records": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Next page cursor may be missing in payload",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("Name")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "records": []
				}`),
			}.Server(),
			Expected:     &common.ReadResult{Done: true, Data: []common.ReadResultRow{}},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is resolved, when provided with a string",
			Input: common.ReadParams{ObjectName: "leads", Fields: connectors.Fields("City")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.QueryParam("q", "SELECT City FROM leads"),
				},
				Then: mockserver.Response(http.StatusOK, responseLeadsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     8,
				NextPage: "/services/data/v60.0/query/01g3A00007lZwLKQA0-2000",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("Department", "AssistantName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.Or{
						mockcond.QueryParam("q", "SELECT AssistantName,Department FROM contacts"),
						mockcond.QueryParam("q", "SELECT Department,AssistantName FROM contacts"),
					},
				},
				Then: mockserver.Response(http.StatusOK, responseListContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 20,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"department":    "Finance",
						"assistantname": "Jean Marie",
					},
					Raw: map[string]any{
						"AccountId":     "001ak00000OKNPHAA5",
						"Department":    "Finance",
						"AssistantName": "Jean Marie",
						"Description":   nil,
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Opportunity with Account association - AccountId added to SOQL and association extracted",
			Input: common.ReadParams{
				ObjectName:        "opportunity",
				Fields:            connectors.Fields("Name", "Amount", "StageName"),
				AssociatedObjects: []string{"account"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.Or{
						// AccountId should be added to the query, order may vary
						mockcond.QueryParam("q", "SELECT Name,Amount,StageName,AccountId FROM opportunity"),
						mockcond.QueryParam("q", "SELECT AccountId,Name,Amount,StageName FROM opportunity"),
						mockcond.QueryParam("q", "SELECT Name,AccountId,Amount,StageName FROM opportunity"),
						mockcond.QueryParam("q", "SELECT Amount,Name,AccountId,StageName FROM opportunity"),
						mockcond.QueryParam("q", "SELECT StageName,Name,Amount,AccountId FROM opportunity"),
					},
				},
				Then: mockserver.Response(http.StatusOK, responseOpportunityWithAccount),
			}.Server(),
			Comparator: comparatorSubsetReadWithAssociations,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Id: "006ak00000OQ4RxAAL",
						Fields: map[string]any{
							"name":      "Test Opportunity 1",
							"amount":    50000.00,
							"stagename": "Prospecting",
						},
						Associations: map[string][]common.Association{
							"account": {
								{
									ObjectId: "001ak00000OKNPHAA5",
									Raw:      nil, // Parent relationships have empty Raw - workflow layer will fetch
								},
							},
						},
						Raw: map[string]any{
							"Id":        "006ak00000OQ4RxAAL",
							"Name":      "Test Opportunity 1",
							"AccountId": "001ak00000OKNPHAA5",
							"Amount":    50000.00,
							"StageName": "Prospecting",
						},
					},
					{
						Id: "006ak00000OQ4RyAAL",
						Fields: map[string]any{
							"name":      "Test Opportunity 2",
							"amount":    30000.00,
							"stagename": "Qualification",
						},
						// No association when AccountId is null
						Associations: nil,
						Raw: map[string]any{
							"Id":        "006ak00000OQ4RyAAL",
							"Name":      "Test Opportunity 2",
							"AccountId": nil,
							"Amount":    30000.00,
							"StageName": "Qualification",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// comparatorSubsetReadWithAssociations extends ComparatorSubsetRead to also validate associations.
func comparatorSubsetReadWithAssociations(serverURL string, actual, expected *common.ReadResult) bool {
	// First check fields, raw, and pagination using the standard comparator
	if !testroutines.ComparatorSubsetRead(serverURL, actual, expected) {
		return false
	}

	// Then check associations
	if len(actual.Data) < len(expected.Data) {
		return false
	}

	for i := range expected.Data {
		if !validateAssociationsForRow(actual.Data[i].Associations, expected.Data[i].Associations) {
			return false
		}
	}

	return true
}

// validateAssociationsForRow validates associations for a single row.
func validateAssociationsForRow(actualAssoc, expectedAssoc map[string][]common.Association) bool {
	// If expected has no associations, actual can have none or some (we don't care)
	if len(expectedAssoc) == 0 {
		return true
	}

	// If expected has associations but actual doesn't, that's a failure
	if len(actualAssoc) == 0 {
		return false
	}

	// Check each expected association type
	for assocType, expectedAssociations := range expectedAssoc {
		actualAssociations, ok := actualAssoc[assocType]
		if !ok {
			return false
		}

		if !validateAssociationsList(actualAssociations, expectedAssociations) {
			return false
		}
	}

	return true
}

// validateAssociationsList validates a list of associations.
func validateAssociationsList(actualAssociations, expectedAssociations []common.Association) bool {
	// Check that we have at least as many associations as expected
	if len(actualAssociations) < len(expectedAssociations) {
		return false
	}

	// Check each expected association
	for j, expectedAssoc := range expectedAssociations {
		actualAssoc := actualAssociations[j]

		if !validateSingleAssociation(actualAssoc, expectedAssoc) {
			return false
		}
	}

	return true
}

// validateSingleAssociation validates a single association.
func validateSingleAssociation(actualAssoc, expectedAssoc common.Association) bool {
	// Check ObjectId
	if expectedAssoc.ObjectId != "" && actualAssoc.ObjectId != expectedAssoc.ObjectId {
		return false
	}

	// Check Raw - if expected is nil, actual should be nil
	if expectedAssoc.Raw == nil && actualAssoc.Raw != nil {
		return false
	}

	// If expected has Raw data, check it matches
	if expectedAssoc.Raw != nil {
		if !reflect.DeepEqual(actualAssoc.Raw, expectedAssoc.Raw) {
			return false
		}
	}

	return true
}

func constructTestConnector(serverURL string) (*Connector, error) {
	return constructTestConnectorGeneral(serverURL, providers.ModuleSalesforceCRM)
}

func constructTestConnectorAccountEngagement(serverURL string) (*Connector, error) {
	return constructTestConnectorGeneral(serverURL, providers.ModuleSalesforceAccountEngagement)
}

func constructTestConnectorGeneral(serverURL string, module common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithWorkspace("test-workspace"),
		WithModule(module),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.moduleInfo.BaseURL, serverURL))

	return connector, nil
}
