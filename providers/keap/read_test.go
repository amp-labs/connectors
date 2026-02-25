package keap

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	millisecondInNano := int(time.Millisecond.Nanoseconds())
	// Errors.
	errorBadRequest := testutils.DataFromFile(t, "get-with-req-body-not-allowed.html")
	errorNotFound := testutils.DataFromFile(t, "url-not-found.html")
	// Version1: Opportunities.
	// responseOpportunitiesModel := testutils.DataFromFile(t, "custom-fields/opportunities-v1.json")
	// responseOpportunities := testutils.DataFromFile(t, "read/opportunities/v1.json")
	// Version 2: Contacts.
	responseContactsModel := testutils.DataFromFile(t, "custom-fields/contacts-v2.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "read/contacts/1-first-page-v2.json")
	responseContactsEmptyPage := testutils.DataFromFile(t, "read/contacts/2-empty-page-v2.json")
	// Tags
	responseTags := testutils.DataFromFile(t, "read/tags/v2.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "messages"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Error cannot send request body on GET operation",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("400 Bad Request: Your client has issued a malformed or illegal request."),
			},
		},
		{
			Name:  "Error page not found",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Keap - Page Not Found"),
			},
		},
		// {
		//	Name: "Opportunities uses custom fields V1",
		//	Input: common.ReadParams{
		//		ObjectName: "opportunities",
		//		Fields: connectors.Fields("opportunity_title",
		//			// Custom fields:
		//			"color"),
		//	},
		//	Server: mockserver.Switch{
		//		Setup: mockserver.ContentJSON(),
		//		Cases: []mockserver.Case{{
		//			If:   mockcond.Path("/crm/rest/v1/opportunities"),
		//			Then: mockserver.Response(http.StatusOK, responseOpportunities),
		//		}, {
		//			If:   mockcond.Path("/crm/rest/v1/opportunities/model"),
		//			Then: mockserver.Response(http.StatusOK, responseOpportunitiesModel),
		//		}},
		//	}.Server(),
		//	Comparator: testroutines.ComparatorSubsetRead,
		//	Expected: &common.ReadResult{
		//		Rows: 1,
		//		Data: []common.ReadResultRow{{
		//			Fields: map[string]any{
		//				"opportunity_title": "First Opportunity",
		//				"color":             "purple",
		//			},
		//			Raw: map[string]any{
		//				"id":           float64(2),
		//				"date_created": "2025-05-28T18:30:05.000+0000",
		//				"last_updated": "2025-05-28T18:30:05.000+0000",
		//				"custom_fields": []any{map[string]any{
		//					"id":      float64(18),
		//					"content": "purple",
		//				}},
		//			},
		//		}},
		//		NextPage: "https://api.infusionsoft.com/crm/rest/v1/opportunities/?limit=1&offset=1000",
		//		Done:     false,
		//	},
		//	ExpectedErrs: nil,
		// },
		{
			Name: "Contacts first page has a link to next",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields: connectors.Fields("given_name",
					// Next fields are custom fields which do NOT exist inside raw.
					// However, they are surfaced to the user via ListObjectMetadata,
					// so they will have context to request them.
					"jobtitle", "jobdescription", "experience", "age"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/crm/rest/v2/contacts"),
					Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
				}, {
					If:   mockcond.Path("/crm/rest/v2/contacts/model"),
					Then: mockserver.Response(http.StatusOK, responseContactsModel),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"given_name":     "Erica",
						"jobtitle":       "Product Owner",
						"jobdescription": "AI application in commerce",
						"experience":     "8 years in 3 companies",
						"age":            float64(32),
					},
					Raw: map[string]any{
						"id":          float64(22),
						"family_name": "Lewis",
						"custom_fields": []any{map[string]any{
							"id":      float64(12),
							"content": "8 years in 3 companies",
						}, map[string]any{
							"id":      float64(6),
							"content": "Product Owner",
						}, map[string]any{
							"id":      float64(8),
							"content": "AI application in commerce",
						}, map[string]any{
							"id":      float64(14),
							"content": float64(32),
						}},
					},
				}, {
					Fields: map[string]any{
						"given_name":     "John",
						"jobtitle":       nil,
						"jobdescription": nil,
						"experience":     nil,
						"age":            nil,
					},
					Raw: map[string]any{
						"id":          float64(20),
						"family_name": "Doe",
					},
				}},
				NextPage: "https://api.infusionsoft.com/crm/rest/v2/contacts/?limit=2&offset=2&since=2024-06-03T22:17:59.039Z&order=id", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of contacts, empty page",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("given_name"),
				Since:      time.Date(2024, 3, 4, 8, 22, 56, 77*millisecondInNano, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/crm/rest/v2/contacts"),
						mockcond.QueryParam("filter", "start_update_time==2024-03-04T08:22:56.077Z"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsEmptyPage),
				}, {
					If:   mockcond.Path("/crm/rest/v2/contacts/model"),
					Then: mockserver.Response(http.StatusOK, []byte{}), // no custom fields
				}},
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Tags page has a link to next",
			Input: common.ReadParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm/rest/v2/tags"),
				Then:  mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "91",
						"name": "Nurture Subscriber",
					},
					Raw: map[string]any{
						"id":          "91",
						"name":        "Nurture Subscriber",
						"description": "",
						"category": map[string]any{
							"id": "10",
						},
					},
					Id: "91",
				}},
				NextPage: "https://api.infusionsoft.com/crm/rest/v2/tags/?page_size=1&page_token=91", // nolint:lll
				Done:     false,
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
