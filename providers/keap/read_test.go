package keap

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

func TestReadV1(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	millisecondInNano := int(time.Millisecond.Nanoseconds())

	errorBadRequest := testutils.DataFromFile(t, "get-with-req-body-not-allowed.html")
	errorNotFound := testutils.DataFromFile(t, "url-not-found.html")
	responseContactsModel := testutils.DataFromFile(t, "custom-fields-contacts.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "read-contacts-1-first-page-v1.json")
	responseContactsEmptyPage := testutils.DataFromFile(t, "read-contacts-2-empty-page-v1.json")

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
				errors.New("400 Bad Request: Your client has issued a malformed or illegal request."), // nolint:goerr113
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
				errors.New("Keap - Page Not Found"), // nolint:goerr113
			},
		},
		{
			Name: "Contacts first page has a link to next",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("given_name", "jobtitle"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/crm/rest/v1/contacts"),
					Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
				}, {
					If:   mockcond.PathSuffix("/crm/rest/v1/contacts/model"),
					Then: mockserver.Response(http.StatusOK, responseContactsModel),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"given_name": "Erica",
						"jobtitle":   "Product Owner",
					},
					Raw: map[string]any{
						"id":             float64(22),
						"family_name":    "Lewis",
						"jobdescription": "AI application in commerce",
						"experience":     "8 years in 3 companies",
						"age":            float64(32),
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
						"given_name": "John",
						"jobtitle":   nil,
					},
					Raw: map[string]any{
						"id":             float64(20),
						"family_name":    "Doe",
						"jobdescription": nil,
						"experience":     nil,
						"age":            nil,
					},
				}},
				NextPage: "https://api.infusionsoft.com/crm/rest/v1/contacts/?limit=2&offset=2&since=2024-06-03T22:17:59.039Z&order=id", // nolint:lll
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
						mockcond.PathSuffix("/crm/rest/v1/contacts"),
						mockcond.QueryParam("since", "2024-03-04T08:22:56.077Z"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsEmptyPage),
				}, {
					If:   mockcond.PathSuffix("/crm/rest/v1/contacts/model"),
					Then: mockserver.Response(http.StatusOK, []byte{}), // no custom fields
				}},
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleV1)
			})
		})
	}
}

func TestReadV2(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseTags := testutils.DataFromFile(t, "read-tags-v2.json")

	tests := []testroutines.Read{
		{
			Name: "Tags page has a link to next",
			Input: common.ReadParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/rest/v2/tags"),
				Then:  mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
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
				return constructTestConnector(tt.Server.URL, ModuleV2)
			})
		})
	}
}

func constructTestConnector(serverURL string, moduleID common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithModule(moduleID),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
