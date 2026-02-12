package salesforce

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSearch(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUnknownObject := testutils.DataFromFile(t, "unknown-object.json")
	responseLeadsFirstPage := testutils.DataFromFile(t, "read-list-leads.json")
	responseListContacts := testutils.DataFromFile(t, "read-list-contacts.json")
	responseOpportunityWithAccount := testutils.DataFromFile(t, "read-opportunity-with-account.json")
	responseOpportunityWithContacts := testutils.DataFromFile(t, "read-opportunity-with-contacts.json")

	tests := []testroutines.Search{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.SearchParams{ObjectName: "Account"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "At least one search filter is requested",
			Input:        common.SearchParams{ObjectName: "Account", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Search with limit is not supported",
			Input: common.SearchParams{
				ObjectName: "Account",
				Fields:     connectors.Fields("id"),
				Filter:     connectors.SearchFilter{}.FilterBy("Name", common.FilterOperatorEQ, "Paris"),
				Limit:      565656,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrPaginationControl},
		},
		{
			Name: "Correct error message is understood from JSON response",
			Input: common.SearchParams{
				ObjectName: "Accout",
				Fields:     connectors.Fields("Name"),
				Filter:     connectors.SearchFilter{}.FilterBy("Name", common.FilterOperatorEQ, "Paris"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseUnknownObject),
			}.Server(),
			ExpectedErrs: []error{
				errors.New("sObject type 'Accout' is not supported"),
			},
		},
		{
			Name: "Next page URL is resolved, when provided with a string",
			Input: common.SearchParams{
				ObjectName: "Leads",
				Fields:     connectors.Fields("Id"),
				Filter:     connectors.SearchFilter{}.FilterBy("Name", common.FilterOperatorEQ, "Paris"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.QueryParam("q", "SELECT Id FROM Leads WHERE Name = 'Paris'"),
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
			Name: "Successful search with chosen fields",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("Department", "AssistantName"),
				Filter:     connectors.SearchFilter{}.FilterBy("Name", common.FilterOperatorEQ, "Paris"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.Permute(
						queryParam("SELECT %v FROM contacts WHERE Name = 'Paris'"),
						"Id", "AssistantName", "Department",
					),
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
			Name: "Search Opportunity with Accounts association - AccountId added to SOQL and association extracted",
			Input: common.SearchParams{
				ObjectName:        "opportunity",
				Fields:            connectors.Fields("Name", "Amount", "StageName"),
				Filter:            connectors.SearchFilter{}.FilterBy("Name", common.FilterOperatorEQ, "Paris"),
				AssociatedObjects: []string{"accounts"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.Permute(
						// Id and AccountId are always added when needed; order may vary.
						queryParam("SELECT %v FROM opportunity WHERE Name = 'Paris'"),
						"Id", "Name", "Amount", "StageName", "AccountId",
					),
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
							"accounts": {
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
		{
			Name: "Search Opportunity with Contacts association via junction - " +
				"OpportunityContactRoles subquery added and ContactIds extracted",
			Input: common.SearchParams{
				ObjectName:        "opportunity",
				Fields:            connectors.Fields("Name", "Amount", "StageName"),
				Filter:            connectors.SearchFilter{}.FilterBy("Name", common.FilterOperatorEQ, "Paris"),
				AssociatedObjects: []string{"contacts"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/services/data/v60.0/query"),
					mockcond.Permute(
						// Id is always added; OpportunityContactRoles subquery order may vary.
						queryParam("SELECT %v FROM opportunity WHERE Name = 'Paris'"),
						"Id", "Name", "Amount", "StageName", "(SELECT FIELDS(STANDARD) FROM OpportunityContactRoles)",
					),
				},
				Then: mockserver.Response(http.StatusOK, responseOpportunityWithContacts),
			}.Server(),
			Comparator: comparatorSubsetReadWithAssociations,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Id: "006ak00000OQ4RxAAL",
						Fields: map[string]any{
							"name":      "Test Opportunity",
							"amount":    50000.00,
							"stagename": "Prospecting",
						},
						Associations: map[string][]common.Association{
							"contacts": {
								{
									ObjectId: "003ak000003dQCGAA2",
									Raw:      nil, // Junction relationships have empty Raw - workflow layer will fetch Contact records
								},
								{
									ObjectId: "003ak000003dQCDAA2",
									Raw:      nil, // Junction relationships have empty Raw - workflow layer will fetch Contact records
								},
							},
						},
						Raw: map[string]any{
							"Id":        "006ak00000OQ4RxAAL",
							"Name":      "Test Opportunity",
							"Amount":    50000.00,
							"StageName": "Prospecting",
							"attributes": map[string]any{
								"type": "Opportunity",
								"url":  "/services/data/v60.0/sobjects/Opportunity/006ak00000OQ4RxAAL",
							},
							"OpportunityContactRoles": map[string]any{
								"totalSize": 2.0,
								"done":      true,
								"records": []any{
									map[string]any{
										"Id":        "00kak00000OQ4RxAAL",
										"ContactId": "003ak000003dQCGAA2",
										"Role":      "Decision Maker",
										"attributes": map[string]any{
											"type": "OpportunityContactRole",
											"url":  "/services/data/v60.0/sobjects/OpportunityContactRole/00kak00000OQ4RxAAL",
										},
									},
									map[string]any{
										"Id":        "00kak00000OQ4RyAAL",
										"ContactId": "003ak000003dQCDAA2",
										"Role":      "Influencer",
										"attributes": map[string]any{
											"type": "OpportunityContactRole",
											"url":  "/services/data/v60.0/sobjects/OpportunityContactRole/00kak00000OQ4RyAAL",
										},
									},
								},
							},
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

			tt.Run(t, func() (connectors.SearchConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
