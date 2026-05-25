package hubspot

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

	requestContactsSince := testutils.DataFromFile(t, "read/objects-api/contacts-req-payload-since.json")
	requestContactsUntil := testutils.DataFromFile(t, "read/objects-api/contacts-req-payload-until.json")
	requestContactsSinceUntil := testutils.DataFromFile(t, "read/objects-api/contacts-req-payload-since-until.json")
	responseContacts := testutils.DataFromFile(t, "read/objects-api/contacts-response.json")
	responseListsFirst := testutils.DataFromFile(t, "read-lists-1-first-page.json")
	responseListsLast := testutils.DataFromFile(t, "read-lists-2-second-page.json")
	responseCampaignsFirst := testutils.DataFromFile(t, "read/campaigns/1-first-page.json")
	responseCampaignsLast := testutils.DataFromFile(t, "read/campaigns/2-last-page.json")
	responseMarketingEmailFirst := testutils.DataFromFile(t, "read/marketing-emails/1-first-page.json")
	responseMarketingEmailLast := testutils.DataFromFile(t, "read/marketing-emails/2-last-page.json")
	responseMarketingForms := testutils.DataFromFile(t, "read/marketing-forms.json")
	responseMarketingEvents := testutils.DataFromFile(t, "read/marketing-events.json")
	responseMeetingLinks := testutils.DataFromFile(t, "read/meeting-links.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Contacts uses object API endpoint",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm/v3/objects/contacts"),
				Then:  mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"email": "a@example.com",
						},
						Id: "1",
						Raw: map[string]any{
							"id": "1",
							"properties": map[string]any{
								"createdate":       "2023-10-26T17:55:48.301Z",
								"email":            "a@example.com",
								"lastmodifieddate": "2024-12-24T17:31:54.727Z",
							},
							"createdAt": "2023-10-26T17:55:48.301Z",
							"updatedAt": "2024-12-24T17:31:54.727Z",
							"archived":  false,
						},
					}, {
						Fields: map[string]any{
							"email": "b@example.com",
						},
						Id: "51",
						Raw: map[string]any{
							"id": "51",
							"properties": map[string]any{
								"createdate":       "2023-10-26T17:55:48.691Z",
								"email":            "b@example.com",
								"lastmodifieddate": "2023-12-13T22:45:30.353Z",
							},
							"createdAt": "2023-10-26T17:55:48.691Z",
							"updatedAt": "2023-12-13T22:45:30.353Z",
							"archived":  false,
						},
					}, {
						Fields: map[string]any{
							"email": "c@example.com",
						},
						Id: "101",
						Raw: map[string]any{
							"id": "101",
							"properties": map[string]any{
								"createdate":       "2023-12-13T22:20:02.649Z",
								"email":            "c@example.com",
								"lastmodifieddate": "2023-12-13T22:20:05.498Z",
							},
							"createdAt": "2023-12-13T22:20:02.649Z",
							"updatedAt": "2023-12-13T22:20:05.498Z",
							"archived":  false,
						},
					},
				},
				NextPage: "https://api.hubapi.com/crm/v3/objects/contacts?limit=100&properties=listId%2Cname&after=394", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contacts records since time",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.BodyBytes(requestContactsSince),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     3,
				NextPage: "394",
				Done:     false,
			},
		},
		{
			Name: "Contacts records until time",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				Until: time.Date(2025, 1, 1, 0, 0, 0, 0,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.BodyBytes(requestContactsUntil),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     3,
				NextPage: "394",
				Done:     false,
			},
		},
		{
			Name: "Contacts records from 'since' till 'until'",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
				Until: time.Date(2025, 1, 1, 0, 0, 0, 0,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/objects/contacts/search"),
					mockcond.BodyBytes(requestContactsSinceUntil),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     3,
				NextPage: "394",
				Done:     false,
			},
		},
		{
			Name: "Lists first page is done via search",
			Input: common.ReadParams{
				ObjectName: "lists",
				Fields:     connectors.Fields("processingType"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/lists/search"),
					mockcond.Body(`{"offset":0,"count":100}`),
				},
				Then: mockserver.Response(http.StatusOK, responseListsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"processingtype": "DYNAMIC",
					},
					Raw: map[string]any{
						// "listId": "3",
						"name": "Test List",
					},
				}, {
					Fields: map[string]any{
						"processingtype": "SNAPSHOT",
					},
					Raw: map[string]any{
						// "listId": "4",
						"name": "Test static company list",
					},
				}},
				NextPage: "2", // Next page token is in fact an offset.
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Lists next page sends offset in payload",
			Input: common.ReadParams{
				ObjectName: "lists",
				Fields:     connectors.Fields("name"),
				NextPage:   "2", // Move offset 2 records ahead to get next page.
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v3/lists/search"),
					mockcond.Body(`{
						"offset": 2,
						"count": 100
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseListsLast),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				Data:     []common.ReadResultRow{},
				NextPage: "", // empty next page is inferred from response
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read marketing campaigns first page",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("hs_name", "hs_notes", "hs_budget_items_sum_amount"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("sort", "-updatedAt"),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaignsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"hs_name":                    "Nurture",
						"hs_notes":                   "Creating campaign from the Dashboard",
						"hs_budget_items_sum_amount": "2.0",
					},
					Raw: map[string]any{
						"id": "84f199fa-beb7-4dca-ad94-3d778cdce157",
						"properties": map[string]any{
							"hs_name":                    "Nurture",
							"hs_notes":                   "Creating campaign from the Dashboard",
							"hs_budget_items_sum_amount": "2.0",
						},
						"createdAt": "2026-05-05T23:41:20.330Z",
						"updatedAt": "2026-05-05T23:45:04.200Z",
					},
					Id: "84f199fa-beb7-4dca-ad94-3d778cdce157",
				}, {
					Fields: map[string]any{
						"hs_name": "Kiwi",
					},
					Raw: map[string]any{
						"id": "fc4583f7-5cfc-4773-8fa4-076cd4f4ae6d",
						"properties": map[string]any{
							"hs_name": "Kiwi",
						},
						"createdAt": "2026-05-05T23:09:27.549Z",
						"updatedAt": "2026-05-05T23:09:27.713Z",
					},
					Id: "fc4583f7-5cfc-4773-8fa4-076cd4f4ae6d",
				}},
				NextPage: "https://api.hubapi.com/marketing/campaigns/2026-03?limit=2&sort=-updatedAt&properties=hs_name%2Chs_notes%2Chs_budget_items_sum_amount&after=Mg%3D%3D",
				Done:     false,
			},
		},
		{
			Name: "Read marketing campaigns with connector side filtering",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("hs_name"),
				// The first item will be returned, last filtered out.
				// Due to the sort order there will be no next page.
				// The record which is excluded has this timestamp: 2026-05-05T23:09:27.713Z
				Since: time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaignsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"hs_name": "Nurture",
						},
						Raw: map[string]any{
							"updatedAt": "2026-05-05T23:45:04.200Z",
						},
						Id: "84f199fa-beb7-4dca-ad94-3d778cdce157",
					},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read marketing campaigns last page",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("hs_name"),
				NextPage:   testroutines.URLTestServer + "/marketing/campaigns/2026-03?limit=2&sort=-updatedAt&properties=hs_name%2Chs_notes%2Chs_budget_items_sum_amount&after=Mg%3D%3D",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
					mockcond.QueryParam("after", "Mg=="),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaignsLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"hs_name": "Inbound",
						},
						Raw: map[string]any{
							"createdAt": "2026-05-05T23:07:11.797Z",
							"updatedAt": "2026-05-05T23:07:12.040Z",
						},
						Id: "5f7bff76-193f-43af-968b-f13c6576ca76",
					},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read marketing emails first page",
			Input: common.ReadParams{
				ObjectName: "marketing/emails",
				Fields:     connectors.Fields("subject"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/emails/2026-03"),
					mockcond.QueryParam("sort", "-updatedAt"),
				},
				Then: mockserver.Response(http.StatusOK, responseMarketingEmailFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"subject": "Behold the latest version of our newsletter!",
					},
					Raw: map[string]any{
						"createdAt":         "2026-05-07T22:59:00.597Z",
						"createdById":       "82226790",
						"emailTemplateMode": "DRAG_AND_DROP",
					},
					Id: "212476546342",
				}, {
					Fields: map[string]any{
						"subject": "Product Launch",
					},
					Raw: map[string]any{
						"createdAt":         "2024-05-29T22:37:35.474Z",
						"createdById":       "62365053",
						"emailTemplateMode": "DRAG_AND_DROP",
					},
					Id: "168871137104",
				}},
				NextPage: "https://api.hubapi.com/marketing/emails/2026-03?limit=3&sort=-updatedAt&after=Mw%3D%3D",
				Done:     false,
			},
		},
		{
			Name: "Read marketing emails last page",
			Input: common.ReadParams{
				ObjectName: "marketing/emails",
				Fields:     connectors.Fields("subject"),
				NextPage:   testroutines.URLTestServer + "/marketing/emails/2026-03?limit=3&sort=-updatedAt&after=Mw%3D%3D",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/emails/2026-03"),
					mockcond.QueryParam("limit", "3"),
					mockcond.QueryParam("after", "Mw=="),
					mockcond.QueryParam("sort", "-updatedAt"),
				},
				Then: mockserver.Response(http.StatusOK, responseMarketingEmailLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"subject": "Your ticket '{{ticket.subject}}' has been received",
						},
						Raw: map[string]any{
							"createdAt":            "2023-12-08T17:47:58.334Z",
							"createdById":          "100",
							"emailCampaignGroupId": "285768335",
							"emailTemplateMode":    "DRAG_AND_DROP",
						},
						Id: "149139108889",
					},
				},
				NextPage: "",
				Done:     true,
			},
		}, {
			Name: "Read marketing forms",
			Input: common.ReadParams{
				ObjectName: "forms",
				Fields:     connectors.Fields("name", "updatedAt"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/marketing/forms/2026-09-beta"),
				Then:  mockserver.Response(http.StatusOK, responseMarketingForms),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"name":      "Tell me about yourself",
							"updatedat": "2026-05-11T17:30:51.442Z",
						},
						Raw: map[string]any{
							"archived": false,
							"formType": "hubspot",
						},
						Id: "591e2731-c869-445d-b422-26f43145e9d2",
					},
				},
				NextPage: "https://api.hubapi.com/marketing/forms/2026-09-beta?limit=1&after=MQ%3D%3D",
				Done:     false,
			},
		}, {
			Name: "Read marketing events",
			Input: common.ReadParams{
				ObjectName: "marketing-events",
				Fields:     connectors.Fields("eventName", "eventOrganizer"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/marketing/marketing-events/2026-03"),
				Then:  mockserver.Response(http.StatusOK, responseMarketingEvents),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"eventname":      "Party",
							"eventorganizer": "Alice",
						},
						Raw: map[string]any{
							"objectId":        "555442196839",
							"externalEventId": "qwe",
							"eventStatus":     "ONGOING",
							"eventStatusV2":   "ongoing",
						},
						Id: "555442196839",
					},
				},
				NextPage: "https://api.hubapi.com/marketing/marketing-events/2026-03?after=NTU1NDQyMTk2ODQw",
				Done:     false,
			},
		}, {
			Name: "Read meeting-links",
			Input: common.ReadParams{
				ObjectName: "meeting-links",
				Fields:     connectors.Fields("name", "slug"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/scheduler/2026-03/meetings/meeting-links"),
				Then:  mockserver.Response(http.StatusOK, responseMeetingLinks),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"slug": "int/public-gathering",
							"name": "Public Gathering",
						},
						Raw: map[string]any{
							"link": "https://meetings.hubspot.com/int/public-gathering",
							"type": "PERSONAL_LINK",
						},
						Id: "12428962",
					},
				},
				NextPage: "https://api.hubapi.com/scheduler/2026-03/meetings/meeting-links?limit=1&after=MQ%3D%3D",
				Done:     false,
			},
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
