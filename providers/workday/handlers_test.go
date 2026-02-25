package workday

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseWorkers := testutils.DataFromFile(t, "read/workers/response.json")
	responseWorkersPaginated := testutils.DataFromFile(t, "read/workers/paginated-response.json")
	responseWorkersEmpty := testutils.DataFromFile(t, "read/workers/empty-response.json")
	responseWorkersWithCustomFields := testutils.DataFromFile(t, "read/workers/with-custom-fields.json")
	responseCustomFieldDefs := testutils.DataFromFile(t, "custom-fields/workers/definitions.json")
	responseEmptyCustomFieldDefs := testutils.DataFromFile(t, "custom-fields/workers/empty-definitions.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "workers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read workers successfully",
			Input: common.ReadParams{
				ObjectName: "workers",
				Fields:     connectors.Fields("id", "descriptor", "primaryWorkEmail", "isManager"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/ccx/api/v1/testTenant/workers"),
						mockcond.QueryParam("limit", "100"),
					},
					Then: mockserver.Response(http.StatusOK, responseWorkers),
				}, {
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseEmptyCustomFieldDefs),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":               "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor":       "Logan McNeil",
						"primaryworkemail": "lmcneil@workday.net",
						"ismanager":        true,
					},
					Raw: map[string]any{
						"id":               "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor":       "Logan McNeil",
						"primaryWorkEmail": "lmcneil@workday.net",
						"isManager":        true,
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read workers with pagination",
			Input: common.ReadParams{
				ObjectName: "workers",
				Fields:     connectors.Fields("id", "descriptor"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/ccx/api/v1/testTenant/workers"),
						mockcond.QueryParam("limit", "100"),
					},
					Then: mockserver.Response(http.StatusOK, responseWorkersPaginated),
				}, {
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseEmptyCustomFieldDefs),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor": "Logan McNeil",
					},
					Raw: map[string]any{
						"id":         "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor": "Logan McNeil",
					},
				}, {
					Fields: map[string]any{
						"id":         "7bc4660a8de249a89e3c6842bfd76645",
						"descriptor": "Jane Smith",
					},
					Raw: map[string]any{
						"id":         "7bc4660a8de249a89e3c6842bfd76645",
						"descriptor": "Jane Smith",
					},
				}},
				NextPage: testroutines.URLTestServer + "/ccx/api/v1/testTenant/workers?limit=100&offset=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read workers with custom PageSize",
			Input: common.ReadParams{
				ObjectName: "workers",
				Fields:     connectors.Fields("id", "descriptor"),
				PageSize:   20,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/ccx/api/v1/testTenant/workers"),
						mockcond.QueryParam("limit", "20"),
					},
					Then: mockserver.Response(http.StatusOK, responseWorkers),
				}, {
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseEmptyCustomFieldDefs),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor": "Logan McNeil",
					},
					Raw: map[string]any{
						"id":         "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor": "Logan McNeil",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read workers empty response",
			Input: common.ReadParams{
				ObjectName: "workers",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/ccx/api/v1/testTenant/workers"),
						mockcond.QueryParam("limit", "100"),
					},
					Then: mockserver.Response(http.StatusOK, responseWorkersEmpty),
				}, {
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseEmptyCustomFieldDefs),
				}},
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
			Name: "Read workers with custom fields resolved",
			Input: common.ReadParams{
				ObjectName: "workers",
				Fields: connectors.Fields("id", "descriptor",
					// custom_field_department_code and custom_field_years_experience are custom fields
					// resolved from Workday web service aliases.
					"custom_field_department_code", "custom_field_years_experience"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/ccx/api/v1/testTenant/workers"),
					Then: mockserver.Response(http.StatusOK, responseWorkersWithCustomFields),
				}, {
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFieldDefs),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor": "Logan McNeil",
						// Custom fields resolved from aliases to human-readable names.
						"custom_field_department_code":  "Engineering",
						"custom_field_years_experience": float64(5),
					},
					// Raw preserves original alias keys untouched.
					Raw: map[string]any{
						"id":               "3aa5550b7fe348b98d7b5741afc65534",
						"descriptor":       "Logan McNeil",
						"primaryWorkEmail": "lmcneil@workday.net",
						"isManager":        true,
						// Original alias keys are preserved in Raw.
						"customField_departmentCode":  "Engineering",
						"customField_yearsExperience": float64(5),
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
