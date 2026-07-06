package hubspot

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSearch(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidObject := testutils.DataFromFile(t, "search/error-invalid-object.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "search/contacts/1-first-page.json")
	responseContactsLastPage := testutils.DataFromFile(t, "search/contacts/2-second-page.json")
	responseContactsToCompanies := testutils.DataFromFile(t, "search/contacts/contacts-to-companies.json")

	tests := []testroutines.TestCaseSearch{
		{
			Name:         "Object name must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.SearchParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "At least one search filter is requested",
			Input:        common.SearchParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Correct error message is understood from JSON response",
			Input: common.SearchParams{
				ObjectName: "butterflies",
				Fields:     connectors.Fields("firstname"),
				Filter:     connectors.SearchFilter{}.FilterBy("firstname", common.FilterOperatorEQ, "Johnnie"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm/objects/2026-03/butterflies/search"),
				Then:  mockserver.Response(http.StatusBadRequest, responseInvalidObject),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("Unable to infer object type from: butterflies"),
			},
		},
		{
			Name: "Successful search for contacts with associations",
			Input: common.SearchParams{
				ObjectName:        "contacts",
				Fields:            connectors.Fields("firstname", "lastname"),
				Filter:            connectors.SearchFilter{}.FilterBy("firstname", common.FilterOperatorEQ, "Johnnie"),
				AssociatedObjects: []string{"companies"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/objects/2026-03/contacts/search"),
						mockcond.PermuteJSONBody(`{
								"limit":200,
								"filterGroups": [{
									"filters": [{
										"propertyName": "firstname",
										"operator": "EQ",
										"value": "Johnnie"
								}]}],
								"properties": [%props]
							}`,
							mockcond.PermuteSlots{{Name: "props", Values: []string{"firstname", "lastname"}}},
						),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/associations/2026-03/contacts/companies/batch/read"),
						mockcond.Body(`{"inputs":[{"id":"501"}]}`), // contact ID
					},
					Then: mockserver.Response(http.StatusOK, responseContactsToCompanies),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "501",
					Fields: map[string]any{
						"firstname": "Johnnie",
						"lastname":  "Heller",
					},
					Associations: map[string][]common.Association{
						"companies": {{
							ObjectId: "29022297485",
							ProviderAssociationMetadata: map[string]any{
								"associationTypes": []map[string]any{
									{
										"category": "HUBSPOT_DEFINED",
										"typeId":   1,
										"label":    "Primary",
									},
									{
										"category": "HUBSPOT_DEFINED",
										"typeId":   279,
									},
								},
							},
						}},
					},
					Raw: map[string]any{
						"id":  "501",
						"url": "https://app.hubspot.com/contacts/44237313/record/0-1/501",
						"properties": map[string]any{
							"firstname":        "Johnnie",
							"lastname":         "Heller",
							"hs_object_id":     "501",
							"createdate":       "2023-12-14T23:31:55.758Z",
							"lastmodifieddate": "2025-01-27T19:53:15.581Z",
							"email":            "johnnie77@hotmail.com",
							"company":          nil,
							"phone":            nil,
							"website":          nil,
						},
					}},
				},
				NextPage: "54",
				Done:     false,
			},
		},
		{
			Name: "Contacts last page with associations",
			Input: common.SearchParams{
				ObjectName:        "contacts",
				Fields:            connectors.Fields("firstname", "lastname"),
				Filter:            connectors.SearchFilter{}.FilterBy("firstname", common.FilterOperatorEQ, "Johnnie"),
				AssociatedObjects: []string{"companies"},
				NextPage:          "54",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/crm/objects/2026-03/contacts/search"),
					Then: mockserver.Response(http.StatusOK, responseContactsLastPage),
				}, {
					If: mockcond.And{
						mockcond.Path("/crm/associations/2026-03/contacts/companies/batch/read"),
						mockcond.Body(`{"inputs":[{"id":"103107104018"}]}`), // contact ID
					},
					Then: mockserver.Response(http.StatusOK, responseContactsToCompanies),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "103107104018",
					Fields: map[string]any{
						"firstname": "Johnnie",
						"lastname":  "Conroy",
					},
					Associations: nil, // No associations for this contact identifier.
					Raw: map[string]any{
						"id":  "103107104018",
						"url": "https://app.hubspot.com/contacts/44237313/record/0-1/103107104018",
						"properties": map[string]any{
							"company":          nil,
							"createdate":       "2025-03-01T19:25:58.039Z",
							"email":            "kaseyweissnat@volkman.biz",
							"firstname":        "Johnnie",
							"hs_object_id":     "103107104018",
							"lastmodifieddate": "2025-03-01T19:26:12.592Z",
							"lastname":         "Conroy",
							"phone":            "9199767082",
							"website":          nil,
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableSearcher, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestCheckSearchResultsLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		nextPage    common.NextPageToken
		expectError bool
	}{
		{
			name:        "Empty next page token is valid",
			nextPage:    "",
			expectError: false,
		},
		{
			name:        "Zero offset is valid",
			nextPage:    "0",
			expectError: false,
		},
		{
			name:        "Offset below limit is valid",
			nextPage:    "8999",
			expectError: false,
		},
		{
			name:        "Offset at limit returns error",
			nextPage:    "10000",
			expectError: true,
		},
		{
			name:        "Offset above limit returns error",
			nextPage:    "10001",
			expectError: true,
		},
		{
			name:        "Large offset returns error",
			nextPage:    "99999",
			expectError: true,
		},
		{
			name:        "Non-numeric token is ignored (allowed to proceed)",
			nextPage:    "not-a-number",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := checkSearchResultsLimit(tt.nextPage)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for nextPage=%q, got nil", tt.nextPage)
				}

				if !errors.Is(err, common.ErrResultsLimitExceeded) {
					t.Errorf("expected ErrResultsLimitExceeded, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for nextPage=%q: %v", tt.nextPage, err)
				}
			}
		})
	}
}

func TestSearchResultsLimitConstant(t *testing.T) {
	t.Parallel()

	// Ensure the limit constant matches HubSpot's documented limit
	if searchResultsLimit != 10000 {
		t.Errorf("searchResultsLimit should be 10000, got %d", searchResultsLimit)
	}
}

func TestReadUsingSearchAPI(t *testing.T) {
	t.Parallel()

	payloadContacts := testutils.DataFromFile(t, "read-via-search/contacts-payload.json")
	responseContacts := testutils.DataFromFile(t, "read-via-search/contacts-response.json")
	responseCampaigns := testutils.DataFromFile(t, "read-via-search/campaigns-response.json")

	contactsOutput := &common.ReadResult{
		Rows: 1,
		Data: []common.ReadResultRow{{
			Id: "220006890315",
			Fields: map[string]any{
				"email": "effieklestz@yahoo.com",
			},
			Raw: map[string]any{
				"id": "220006890315",
				"properties": map[string]any{
					"company":          nil,
					"createdate":       "2026-05-06T11:48:03.785Z",
					"email":            "effieklestz@yahoo.com",
					"firstname":        nil,
					"hs_object_id":     "220006890315",
					"lastmodifieddate": "2026-05-06T19:26:18.036Z",
					"lastname":         nil,
					"phone":            nil,
					"website":          nil,
				},
				"createdAt": "2026-05-06T11:48:03.785Z",
				"updatedAt": "2026-05-06T19:26:18.036Z",
				"archived":  false,
				"url":       "https://app.hubspot.com/contacts/44623425/record/0-1/220006890315",
			},
		}},
		Done: true,
	}

	tests := []SearchViaRead{
		{
			Name: "Read contacts using search API without filters",
			Input: SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.BodyBytes(payloadContacts),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Expected: contactsOutput,
		},
		{
			Name: "Read contacts using search API with filters and since",
			Input: SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
				FilterGroups: []FilterGroup{{
					Filters: []Filter{
						BuildLastModifiedFilterGroup(&common.ReadParams{
							ObjectName: "contacts",
							Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
						}),
					},
				}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.BodyBytes(payloadContacts),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Expected: contactsOutput,
		},
		{
			Name: "Read contacts using search API with filters and no since",
			Input: SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				FilterGroups: []FilterGroup{{
					Filters: []Filter{
						BuildLastModifiedFilterGroup(&common.ReadParams{
							ObjectName: "contacts",
							Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
						}),
					},
				}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.BodyBytes(payloadContacts),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Expected: contactsOutput,
		},
		{
			Name: "Read marketing campaigns via Search",
			Input: SearchParams{
				ObjectName: "marketing-campaigns",
				Fields:     connectors.Fields("hs_budget_items_sum_amount", "hs_name", "hs_notes"),
				Since:      time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
					mockcond.BodyBytes(nil),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaigns),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "84f199fa-beb7-4dca-ad94-3d778cdce157",
					Fields: map[string]any{
						"hs_budget_items_sum_amount": "2.0",
						"hs_name":                    "Nurture",
						"hs_notes":                   "Creating campaign from the Dashboard",
					},
					Raw: map[string]any{
						"id": "84f199fa-beb7-4dca-ad94-3d778cdce157",
						"properties": map[string]any{
							"hs_budget_items_sum_amount": "2.0",
							"hs_name":                    "Nurture",
							"hs_notes":                   "Creating campaign from the Dashboard",
						},
						"createdAt": "2026-05-05T23:41:20.330Z",
						"updatedAt": "2026-05-05T23:45:04.200Z",
						"businessUnits": []any{
							map[string]any{"id": float64(0)},
						},
					},
				}},
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (*Connector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

type (
	SearchViaReadType = testroutines.TestCase[SearchParams, *common.ReadResult]
	SearchViaRead     SearchViaReadType
)

func (r SearchViaRead) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()

	t.Cleanup(func() {
		SearchViaReadType(r).Close()
	})

	conn := builder.Build(t, r.Name)
	output, err := conn.ReadUsingSearchAPI(t.Context(), r.Input)

	SearchViaReadType(r).Validate(t, err, output)
}
