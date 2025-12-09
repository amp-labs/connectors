package sellsy

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorBadRequest := testutils.DataFromFile(t, "read/err-bad-request.json")
	responseContactsFirstPage := testutils.DataFromFile(t, "read/contacts/1-first-page.json")
	responseContactsLastPage := testutils.DataFromFile(t, "read/contacts/2-last-page.json")
	responseContactsEmptyPage := testutils.DataFromFile(t, "read/contacts/empty.json")
	responseTasksLabelsFirstPage := testutils.DataFromFile(t, "read/tasks-labels/1-first-page.json")

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
			Name:  "Error invalid params",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("last_name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/custom-fields"),
				Then:  mockserver.Response(http.StatusOK, nil),
				Else:  mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Le contenu de la requête est invalide: le champ 'filters' est manquant."), // nolint:lll
			},
		},
		{
			Name: "Read contacts first page via search endpoint with payload",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("last_name"),
				Since:      time.Date(2025, 8, 22, 8, 22, 56, 0, time.UTC),
				Until:      time.Date(2025, 8, 25, 8, 32, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v2/contacts/search"),
						mockcond.Body(`{
						"filters": {
							"updated": {
								"start":"2025-08-22T08:22:56Z",
								"end":"2025-08-25T08:32:00Z"
							}
						}
					}`),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
				}, {
					If:   mockcond.Path("/v2/custom-fields"),
					Then: mockserver.Response(http.StatusOK, nil),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"last_name": "Blanc",
					},
					Raw: map[string]any{
						"civility":      "mr",
						"first_name":    "Antoine",
						"email":         "antoine@sellsy-mail.com",
						"mobile_number": "+33612345678",
					},
				}, {
					Fields: map[string]any{
						"last_name": "Leduc",
					},
					Raw: map[string]any{
						"civility":      "mrs",
						"first_name":    "Marie",
						"email":         "marie@sellsy-mail.com",
						"mobile_number": "+33612345678",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/contacts/search?limit=100&offset=WyI0Il0=",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts second page via search endpoint using next page token",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("last_name"),
				NextPage:   testroutines.URLTestServer + "/v2/contacts/search?limit=100&offset=WyI0Il0=",
				Since:      time.Date(2025, 8, 22, 8, 22, 56, 0, time.UTC),
				Until:      time.Date(2025, 8, 25, 8, 32, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v2/contacts/search"),
						mockcond.QueryParam("limit", "100"),
						mockcond.QueryParam("offset", "WyI0Il0="),
						mockcond.Body(`{
						"filters": {
							"updated": {
								"start":"2025-08-22T08:22:56Z",
								"end":"2025-08-25T08:32:00Z"
							}
						}
					}`),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsLastPage),
				}, {
					If:   mockcond.Path("/v2/custom-fields"),
					Then: mockserver.Response(http.StatusOK, nil),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"last_name": "Durand",
					},
					Raw: map[string]any{
						"civility":      "mr",
						"first_name":    "Michel",
						"email":         "michel@sellsy-enchante.fr",
						"mobile_number": "+33612345678",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts empty page does not produce next page token",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("last_name"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v2/contacts/search"),
						mockcond.Body(`{
						"filters": {
							"updated": {}
						}
					}`),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsEmptyPage),
				}, {
					If:   mockcond.Path("/v2/custom-fields"),
					Then: mockserver.Response(http.StatusOK, nil),
				}},
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
			Name: "Read task labels via usual get endpoint",
			Input: common.ReadParams{
				ObjectName: "tasks/labels",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/tasks/labels"),
					mockcond.BodyBytes(nil),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksLabelsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "📞 Relance téléphonique",
					},
					Raw: map[string]any{
						"id":        float64(3),
						"name":      "📞 Relance téléphonique",
						"color":     "FAE60C",
						"is_active": true,
						"rank":      float64(0),
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/tasks/labels?limit=100&offset=WyIwIiwiMyJd",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read task labels last page via usual get endpoint",
			Input: common.ReadParams{
				ObjectName: "tasks/labels",
				Fields:     connectors.Fields("name"),
				NextPage:   testroutines.URLTestServer + "/v2/tasks/labels?limit=100&offset=WyIwIiwiMyJd",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v2/tasks/labels"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("offset", "WyIwIiwiMyJd"),
					mockcond.BodyBytes(nil),
				},
				// Response doesn't matter for second page, reuse first page output.
				Then: mockserver.Response(http.StatusOK, responseTasksLabelsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				Data:     []common.ReadResultRow{},
				NextPage: testroutines.URLTestServer + "/v2/tasks/labels?limit=100&offset=WyIwIiwiMyJd",
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
