package salesforce

import (
	"errors"
	"net/http"
	"testing"
	"time"

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

func TestReadPardot(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseMissingHeader := testutils.DataFromFile(t, "pardot/read/emails/err-missing-header.json")
	responseMissingQuery := testutils.DataFromFile(t, "pardot/read/emails/err-missing-query.json")
	responseEmailsFirstPage := testutils.DataFromFile(t, "pardot/read/emails/1-first-page.json")
	responseEmailsEmptyPage := testutils.DataFromFile(t, "pardot/read/emails/2-empty-page.json")

	pardotHeader := http.Header{
		"Pardot-Business-Unit-Id": []string{"test-business-unit-id"},
	}

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "emails"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error Missing Header message is understood",
			Input: common.ReadParams{ObjectName: "emails", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseMissingHeader),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("A required header is missing: Pardot-Business-Unit-Id header not found on request."), // nolint:goerr113
			},
		},
		{
			Name:  "Error Missing Query message is understood",
			Input: common.ReadParams{ObjectName: "emails", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseMissingQuery),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("One or more required parameters are missing: fields"), // nolint:goerr113
			},
		},
		{
			Name: "Read emails first page",
			Input: common.ReadParams{
				ObjectName: "eMaILs",
				Fields:     connectors.Fields("name", "subject"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v5/objects/emails"),
					mockcond.QueryParam("limit", "1000"),
					mockcond.Or{
						mockcond.QueryParam("fields", "name,subject"),
						mockcond.QueryParam("fields", "subject,name"),
					},
					mockcond.Header(pardotHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseEmailsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":    "Sending first email ever",
						"subject": "Few Moments Later",
					},
					Raw: map[string]any{
						"id":         float64(34277860),
						"clientType": "Web",
						"updatedAt":  "2025-05-16T11:27:08-07:00",
					},
				}, {
					Fields: map[string]any{
						"name":    "Second email, on the roll",
						"subject": "Second email, unbelievable",
					},
					Raw: map[string]any{
						"id":         float64(34277863),
						"clientType": "Web",
						"updatedAt":  "2025-05-16T11:29:03-07:00",
					},
				}},
				NextPage: "https://pi.demo.pardot.com/api/v5/objects/emails?fields=id,name,subject,clientType,prospectId,listEmailId,updatedAt&nextPageToken=eyJvcmRlckJ5IjoiIiwiZmlsdGVycyI6W10sImxpbWl0IjoxLCJyZXN1bWVWYWx1ZSI6eyJpZCI6MzQyNzc4NjN9LCJwYWdlIjoyLCJyZWNDb3VudCI6MiwiZXhwaXJlVGltZSI6IjIwMjUtMDUtMTZUMjA6Mzc6MTMtMDc6MDAiLCJkZWxldGVkIjpudWxsfQ==", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of emails last empty page",
			Input: common.ReadParams{
				ObjectName: "eMaILs",
				Fields:     connectors.Fields("name"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v5/objects/emails"),
					mockcond.QueryParam("limit", "1000"),
					mockcond.QueryParam("fields", "name"),
					mockcond.QueryParam("sentAtAfterOrEqualTo", "2024-09-19T04:30:45-08:00"),
					mockcond.Header(pardotHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseEmailsEmptyPage),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read Emails using next page token",
			Input: common.ReadParams{
				ObjectName: "eMaILs",
				Fields:     connectors.Fields("name"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
				NextPage: testroutines.URLTestServer + "/api/v5/objects/emails?fields=id,name,subject,clientType,prospectId,listEmailId,updatedAt&nextPageToken=eyJvcmRlckJ5IjoiIiwiZmlsdGVycyI6W10sImxpbWl0IjoxLCJyZXN1bWVWYWx1ZSI6eyJpZCI6MzQyNzc4NjN9LCJwYWdlIjoyLCJyZWNDb3VudCI6MiwiZXhwaXJlVGltZSI6IjIwMjUtMDUtMTZUMjA6Mzc6MTMtMDc6MDAiLCJkZWxldGVkIjpudWxsfQ==", // nolint:lll
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v5/objects/emails"),
					// Provider API doesn't allow these query parameters alongside with nextPageToken.
					mockcond.QueryParamsMissing("sentAtAfterOrEqualTo", "limit"),
					mockcond.Header(pardotHeader),
				},
				Then: mockserver.Response(http.StatusOK, responseEmailsEmptyPage),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
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
				return constructTestConnectorAccountEngagement(tt.Server.URL)
			})
		})
	}
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
		WithMetadata(map[string]string{
			"isDemo":         "true",
			"businessUnitId": "test-business-unit-id",
		}),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.moduleInfo.BaseURL, serverURL))

	return connector, nil
}
