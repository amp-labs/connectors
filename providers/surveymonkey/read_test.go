package surveymonkey

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseGroups := testutils.DataFromFile(t, "read/groups.json")
	responseGroupsPage1 := testutils.DataFromFile(t, "read/groups-page1.json")
	responseGroupsEmpty := testutils.DataFromFile(t, "read/groups-empty.json")
	responseContacts := testutils.DataFromFile(t, "read/contacts.json")
	responseSurveys := testutils.DataFromFile(t, "read/surveys.json")
	responseContactFields := testutils.DataFromFile(t, "read/contact-fields.json")
	responseWorkgroups := testutils.DataFromFile(t, "read/workgroups.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectGroups},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "unknown", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Zero records response for groups",
			Input: common.ReadParams{ObjectName: objectGroups, Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/groups"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseGroupsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read groups",
			Input: common.ReadParams{ObjectName: objectGroups, Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/groups"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseGroups),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "1231511",
						"name": "Your Group",
					},
					Raw: map[string]any{
						"id":   "1231511",
						"name": "Your Group",
						"href": "https://api.surveymonkey.com/v3/groups/1231511",
					},
					Id: "1231511",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read groups with next page from links.next",
			Input: common.ReadParams{ObjectName: objectGroups, Fields: connectors.Fields("id"), PageSize: 1},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/groups"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseGroupsPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "1231511",
					},
					Raw: map[string]any{
						"id":   "1231511",
						"name": "Your Group",
						"href": "https://api.surveymonkey.com/v3/groups/1231511",
					},
					Id: "1231511",
				}},
				NextPage: "https://api.surveymonkey.com/v3/groups?per_page=1&page=2",
				Done:     false,
			},
		},
		{
			Name:  "Read contacts",
			Input: common.ReadParams{ObjectName: objectContacts, Fields: connectors.Fields("id", "email", "first_name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/contacts"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "1234",
						"email":      "jane@example.com",
						"first_name": "Jane",
					},
					Raw: map[string]any{
						"id":           "1234",
						"first_name":   "Jane",
						"last_name":    "Doe",
						"email":        "jane@example.com",
						"phone_number": "",
						"href":         "https://api.surveymonkey.com/v3/contacts/1234",
						"status":       "active",
					},
					Id: "1234",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read surveys",
			Input: common.ReadParams{ObjectName: objectSurveys, Fields: connectors.Fields("id", "title", "owner")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/surveys"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseSurveys),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "1234",
						"title": "My Survey",
						"owner": "5678",
					},
					Raw: map[string]any{
						"id":       "1234",
						"title":    "My Survey",
						"nickname": "",
						"owner":    "5678",
						"href":     "https://api.surveymonkey.com/v3/surveys/1234",
					},
					Id: "1234",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read contact fields",
			Input: common.ReadParams{ObjectName: objectContactFields, Fields: connectors.Fields("id", "label")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/contact_fields"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactFields),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "1",
						"label": "Custom 1",
					},
					Raw: map[string]any{
						"id":    "1",
						"label": "Custom 1",
						"href":  "https://api.surveymonkey.com/v3/contact_fields/1",
					},
					Id: "1",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read workgroups",
			Input: common.ReadParams{ObjectName: objectWorkgroups, Fields: connectors.Fields("id", "name", "organization_id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/workgroups"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseWorkgroups),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":              "9bccda1faa0d45363edb7fc22",
						"name":            "Brand Awareness",
						"organization_id": "2112",
					},
					Raw: map[string]any{
						"id":              "9bccda1faa0d45363edb7fc22",
						"name":            "Brand Awareness",
						"description":     "Spreading the company brand",
						"is_visible":      true,
						"organization_id": "2112",
						"share_count":     "0",
						"members_count":   float64(1),
						"created_at":      "2022-12-01T18:03:45",
						"updated_at":      "2022-12-05T18:03:45",
					},
					Id: "9bccda1faa0d45363edb7fc22",
				}},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
