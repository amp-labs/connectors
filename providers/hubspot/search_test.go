package hubspot

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

func TestSearch(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidObject := testutils.DataFromFile(t, "search/error-invalid-object.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "search/contacts/1-first-page.json")
	responseContactsLastPage := testutils.DataFromFile(t, "search/contacts/2-second-page.json")
	responseContactsToCompanies := testutils.DataFromFile(t, "search/contacts/contacts-to-companies.json")

	tests := []testroutines.Search{
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
				If:    mockcond.Path("/crm/v3/objects/butterflies/search"),
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
						mockcond.Path("/crm/v3/objects/contacts/search"),
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
							mockcond.PermuteSlot{Name: "props", Values: []string{"firstname", "lastname"}},
						),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/v4/associations/contacts/companies/batch/read"),
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
							ObjectId:        "29022297485",
							AssociationType: "category=HUBSPOT_DEFINED id=1 label=Primary",
						}, {
							ObjectId:        "29022297485",
							AssociationType: "category=HUBSPOT_DEFINED id=279",
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
					If:   mockcond.Path("/crm/v3/objects/contacts/search"),
					Then: mockserver.Response(http.StatusOK, responseContactsLastPage),
				}, {
					If: mockcond.And{
						mockcond.Path("/crm/v4/associations/contacts/companies/batch/read"),
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

			tt.Run(t, func() (connectors.SearchConnector, error) {
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
