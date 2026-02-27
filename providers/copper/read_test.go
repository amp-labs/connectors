package copper

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

	errorNotFound := testutils.DataFromFile(t, "not-found.json")
	errorInvalidParams := testutils.DataFromFile(t, "invalid-params.json")
	responseProjectsFirstPage := testutils.DataFromFile(t, "read/projects/1-first-page.json")
	responseProjectsLastPage := testutils.DataFromFile(t, "read/projects/2-last-page.json")
	responseCompanies := testutils.DataFromFile(t, "read/companies/with-custom-fields.json")
	responseCustomFields := testutils.DataFromFile(t, "custom/fields.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "projects"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error endpoint for object is not found",
			Input: common.ReadParams{ObjectName: "projects", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				testutils.StringError("not_found"),
			},
		},
		{
			Name:  "Error invalid params",
			Input: common.ReadParams{ObjectName: "projects", Fields: connectors.Fields("summary")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidParams),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Invalid input: Validation errors: Base: Unrecognized attributes specified: pageSize"), // nolint:lll
			},
		},
		{
			Name: "Read projects first page",
			Input: common.ReadParams{
				ObjectName: "projects",
				Fields:     connectors.Fields("name"),
				Since:      time.Unix(1754518014, 0),
				Until:      time.Unix(1754518016, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/developer_api/v1/projects/search"),
					mockcond.Body(`{
						"sort_by":"date_modified","sort_direction":"desc",
						"minimum_modified_date":"1754518014",
						"maximum_modified_date":"1754518016",
						"page_number":"1","page_size":200}`),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseProjectsFirstPage),
				Else: mockserver.Response(http.StatusOK, responseCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "New Demo Project",
					},
					Raw: map[string]any{
						"date_created":  float64(1754429285),
						"date_modified": float64(1754429285),
					},
				}, {
					Fields: map[string]any{
						"name": "Second project",
					},
					Raw: map[string]any{
						"date_created":  float64(1754429427),
						"date_modified": float64(1754429427),
					},
				}},
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read projects second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "projects",
				Fields:     connectors.Fields("name"),
				NextPage:   "2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/developer_api/v1/projects/search"),
					mockcond.Body(`{
						"sort_by":"date_modified","sort_direction":"desc",
						"page_number":"2","page_size":200}`),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseProjectsLastPage),
				Else: mockserver.Response(http.StatusOK, responseCustomFields),
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
			Name: "Read company where custom fields are resolved",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields: connectors.Fields("name",
					"custom_field_birthday", "custom_field_fruits", "custom_field_isbouillonsoup"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/developer_api/v1/companies/search"),
						mockcond.Body(`{
						"sort_by":"date_modified","sort_direction":"desc",
						"page_number":"1","page_size":200}`),
						mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
						mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseCompanies),
				}, {
					If: mockcond.And{
						mockcond.Path("/developer_api/v1/custom_field_definitions"),
						mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
						mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":                        "Demo Company",
						"custom_field_birthday":       float64(1754031600),
						"custom_field_fruits":         float64(2082340),
						"custom_field_isbouillonsoup": true,
					},
					Raw: map[string]any{
						"details": "This is a demo company",
					},
				}},
				NextPage: "2",
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
