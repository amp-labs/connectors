package insightly

import (
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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseLeadsFirstPage := testutils.DataFromFile(t, "read/leads/1-first-page.json")
	responseLeadsSecondPage := testutils.DataFromFile(t, "read/leads/2-second-page.json")
	responseLeadsLastPage := testutils.DataFromFile(t, "read/leads/3-last-page.json")
	responseContacts := testutils.DataFromFile(t, "read/contacts/list.json")
	responseFruits := testutils.DataFromFile(t, "read/fruits-custom-object/list.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
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
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "CommunityComments", Fields: connectors.Fields("BODY")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentText(),
				Always: mockserver.ResponseString(http.StatusNotFound,
					"API User does not have access to CommunityComments."),
			}.Server(),
			ExpectedErrs: []error{
				testutils.StringError("API User does not have access to CommunityComments."),
				common.ErrRetryable,
			},
		},
		{
			Name: "Read leads first page",
			Input: common.ReadParams{
				ObjectName: "Leads",
				Fields:     connectors.Fields("FIRST_NAME", "LAST_NAME"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3.1/Leads/Search"),
					mockcond.QueryParam("top", "500"),
					mockcond.QueryParamsMissing("skip"),
				},
				Then: mockserver.Response(http.StatusOK, responseLeadsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"first_name": "Katherine",
						"last_name":  "Nguyen",
					},
					Raw: map[string]any{
						"ADDRESS_COUNTRY": "United States",
						"TAGS": []any{map[string]any{
							"TAG_NAME": "Paris",
						}},
					},
				}, {
					Fields: map[string]any{
						"first_name": "Miquel",
						"last_name":  "Anthony",
					},
					Raw: map[string]any{
						"ADDRESS_COUNTRY": "United States",
						"TAGS": []any{map[string]any{
							"TAG_NAME": "Washington",
						}},
					},
				}},
				NextPage: testroutines.URLTestServer + "/v3.1/Leads/Search?skip=500&top=500",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read leads second page uses NextPage token",
			Input: common.ReadParams{
				ObjectName: "Leads",
				Fields:     connectors.Fields("FIRST_NAME", "LAST_NAME"),
				NextPage:   testroutines.URLTestServer + "/v3.1/Leads/Search?skip=500&top=500",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3.1/Leads/Search"),
					mockcond.QueryParam("top", "500"),
					mockcond.QueryParam("skip", "500"),
				},
				Then: mockserver.Response(http.StatusOK, responseLeadsSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"first_name": "Fred",
						"last_name":  "Everett",
					},
					Raw: map[string]any{
						"ADDRESS_COUNTRY": "United States",
						"TAGS": []any{map[string]any{
							"TAG_NAME": "Warsaw",
						}},
					},
				}},
				NextPage: testroutines.URLTestServer + "/v3.1/Leads/Search?top=500&skip=1000",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read leads last page",
			Input: common.ReadParams{
				ObjectName: "Leads",
				Fields:     connectors.Fields("FIRST_NAME", "LAST_NAME"),
				NextPage:   testroutines.URLTestServer + "/v3.1/Leads/Search?top=500&skip=1000",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3.1/Leads/Search"),
					mockcond.QueryParam("top", "500"),
					mockcond.QueryParam("skip", "1000"),
				},
				Then: mockserver.Response(http.StatusOK, responseLeadsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Leads incremental read",
			Input: common.ReadParams{
				ObjectName: "Leads",
				Fields:     connectors.Fields("FIRST_NAME", "LAST_NAME"),
				Since:      time.Date(2024, 3, 4, 8, 22, 56, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3.1/Leads/Search"),
					mockcond.QueryParam("top", "500"),
					mockcond.QueryParamsMissing("skip"),
					mockcond.QueryParam("updated_after_utc", "2024-03-04T08:22:56Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseLeadsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				NextPage: testroutines.URLTestServer +
					"/v3.1/Leads/Search?skip=500&top=500&updated_after_utc=2024-03-04T08:22:56Z",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contacts with custom fields",
			Input: common.ReadParams{
				ObjectName: "Contacts",
				Fields:     connectors.Fields("CONTACT_ID", "EMAIL_ADDRESS", "Hobby__c", "Interests__c"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3.1/Contacts/Search"),
				Then:  mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email_address": "pamela@mail.com",
						"contact_id":    float64(366638973),
						"hobby__c":      "Skiing",
						"interests__c":  "Music;Sports;Travel",
					},
					Raw: map[string]any{
						"FIRST_NAME": "Pamela",
						"CUSTOMFIELDS": []any{
							map[string]any{
								"FIELD_NAME":      "Hobby__c",
								"FIELD_VALUE":     "Skiing",
								"CUSTOM_FIELD_ID": "Hobby__c",
							},
							map[string]any{
								"FIELD_NAME":      "Interests__c",
								"FIELD_VALUE":     "Music;Sports;Travel",
								"CUSTOM_FIELD_ID": "Interests__c",
							},
							map[string]any{
								"FIELD_NAME":      "Newsletter_Subscription__c",
								"FIELD_VALUE":     true,
								"CUSTOM_FIELD_ID": "Newsletter_Subscription__c",
							},
							map[string]any{
								"FIELD_NAME":      "Preferred_Contact_Method__c",
								"FIELD_VALUE":     "SMS",
								"CUSTOM_FIELD_ID": "Preferred_Contact_Method__c",
							},
						},
					},
				}},
				NextPage: testroutines.URLTestServer +
					"/v3.1/Contacts/Search?skip=500&top=500",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Custom object fruits with custom fields",
			Input: common.ReadParams{
				ObjectName: "Fruit__c",
				Fields:     connectors.Fields("RECORD_NAME", "RECORD_ID", "Color__c", "Weight__c"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3.1/Fruit__c/Search"),
				Then:  mockserver.Response(http.StatusOK, responseFruits),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"record_id":   float64(54840676),
						"record_name": "Apple",
						"color__c":    "Green",
						"weight__c":   float64(12),
					},
					Raw: map[string]any{
						"CUSTOMFIELDS": []any{
							map[string]any{
								"FIELD_NAME":      "Color__c",
								"FIELD_VALUE":     "Green",
								"CUSTOM_FIELD_ID": "Color__c",
							},
							map[string]any{
								"FIELD_NAME":      "Weight__c",
								"FIELD_VALUE":     float64(12),
								"CUSTOM_FIELD_ID": "Weight__c",
							},
						},
					},
				}, {
					Fields: map[string]any{
						"record_id":   float64(54840682),
						"record_name": "Banana",
						"color__c":    nil,
						"weight__c":   3.2,
					},
					Raw: map[string]any{
						"CUSTOMFIELDS": []any{
							map[string]any{
								"FIELD_NAME":      "Color__c",
								"FIELD_VALUE":     nil,
								"CUSTOM_FIELD_ID": "Color__c",
							},
							map[string]any{
								"FIELD_NAME":      "Weight__c",
								"FIELD_VALUE":     3.2,
								"CUSTOM_FIELD_ID": "Weight__c",
							},
						},
					},
				}},
				NextPage: testroutines.URLTestServer +
					"/v3.1/Fruit__c/Search?skip=500&top=500",
				Done: false,
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
