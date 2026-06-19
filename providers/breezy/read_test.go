package breezy

import (
	"errors"
	"net/http"
	"sort"
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

const testCompanyID = "testCompanyID"

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseCompanies := testutils.DataFromFile(t, "read/companies.json")
	responsePositions := testutils.DataFromFile(t, "read/positions.json")
	responsePipelines := testutils.DataFromFile(t, "read/pipelines.json")
	responseCategories := testutils.DataFromFile(t, "read/categories.json")
	responseDepartments := testutils.DataFromFile(t, "read/departments.json")
	responseQuestionnaires := testutils.DataFromFile(t, "read/questionnaires.json")
	responseTemplates := testutils.DataFromFile(t, "read/templates.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectCompanies},
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
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: objectDepartments, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/departments"),
				},
				Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read/empty-array.json")),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Read positions with Since after all records returns empty",
			Input: common.ReadParams{
				ObjectName: objectPositions,
				Fields:     connectors.Fields("_id", "name"),
				Since:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/positions"),
				},
				Then: mockserver.Response(http.StatusOK, responsePositions),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read companies",
			Input: common.ReadParams{ObjectName: objectCompanies, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/companies"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "abc123def456",
						"name": "Acme Corp",
					},
					Raw: map[string]any{
						"_id":          "abc123def456",
						"name":         "Acme Corp",
						"friendly_id":  "acme",
						"initial":      "A",
						"member_count": float64(5),
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read positions",
			Input: common.ReadParams{ObjectName: objectPositions, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/positions"),
				},
				Then: mockserver.Response(http.StatusOK, responsePositions),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "pos001",
						"name": "Software Engineer",
					},
					Raw: map[string]any{
						"_id":          "pos001",
						"name":         "Software Engineer",
						"type":         "fullTime",
						"state":        "published",
						"department":   "Engineering",
						"updated_date": "2024-06-01T10:00:00Z",
					},
				}, {
					Fields: map[string]any{
						"_id":  "pos002",
						"name": "Product Manager",
					},
					Raw: map[string]any{
						"_id":          "pos002",
						"name":         "Product Manager",
						"type":         "fullTime",
						"state":        "published",
						"department":   "Product",
						"updated_date": "2024-06-15T12:00:00Z",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read positions with Since filters connector-side",
			Input: common.ReadParams{
				ObjectName: objectPositions,
				Fields:     connectors.Fields("_id", "name"),
				Since:      time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/positions"),
				},
				Then: mockserver.Response(http.StatusOK, responsePositions),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "pos002",
						"name": "Product Manager",
					},
					Raw: map[string]any{
						"_id":          "pos002",
						"name":         "Product Manager",
						"updated_date": "2024-06-15T12:00:00Z",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read pipelines (map response)",
			Input: common.ReadParams{ObjectName: objectPipelines, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/pipelines"),
				},
				Then: mockserver.Response(http.StatusOK, responsePipelines),
			}.Server(),
			Comparator: comparatorSubsetReadOrderByID,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "default",
						"name": "Default Pipeline",
					},
					Raw: map[string]any{
						"_id":      "default",
						"name":     "Default Pipeline",
						"type":     "default",
						"pipeline": []any{},
					},
				}, {
					Fields: map[string]any{
						"_id":  "default_pool",
						"name": "Default Pool",
					},
					Raw: map[string]any{
						"_id":      "default_pool",
						"name":     "Default Pool",
						"type":     "pool",
						"pipeline": []any{},
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read categories",
			Input: common.ReadParams{ObjectName: objectCategories, Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/categories"),
				},
				Then: mockserver.Response(http.StatusOK, responseCategories),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "software",
						"name": "Software Development",
					},
					Raw: map[string]any{
						"id":   "software",
						"name": "Software Development",
					},
				}, {
					Fields: map[string]any{
						"id":   "design",
						"name": "Interactive Design",
					},
					Raw: map[string]any{
						"id":   "design",
						"name": "Interactive Design",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read departments",
			Input: common.ReadParams{ObjectName: objectDepartments, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/departments"),
				},
				Then: mockserver.Response(http.StatusOK, responseDepartments),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "dept001",
						"name": "Engineering",
					},
					Raw: map[string]any{
						"_id":  "dept001",
						"name": "Engineering",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read questionnaires",
			Input: common.ReadParams{ObjectName: objectQuestionnaires, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/questionnaires"),
				},
				Then: mockserver.Response(http.StatusOK, responseQuestionnaires),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "q001",
						"name": "Test 1",
					},
					Raw: map[string]any{
						"_id":  "q001",
						"name": "Test 1",
						"sections": []any{},
						"message_template": map[string]any{
							"body": "Hi [[candidate_first_name]]",
							"type": "message",
						},
						"options": map[string]any{
							"move_to": map[string]any{},
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read templates",
			Input: common.ReadParams{ObjectName: objectTemplates, Fields: connectors.Fields("_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/company/" + testCompanyID + "/templates"),
				},
				Then: mockserver.Response(http.StatusOK, responseTemplates),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"_id":  "tpl001",
						"name": "Welcome Email",
					},
					Raw: map[string]any{
						"_id":  "tpl001",
						"name": "Welcome Email",
						"body": "Hello [[candidate_first_name]]",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestNewConnectorRequiresCompanyIDMetadata(t *testing.T) {
	t.Parallel()

	_, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if !errors.Is(err, common.ErrMissingMetadata) {
		t.Fatalf("expected ErrMissingMetadata, got %v", err)
	}
}

func comparatorSubsetReadOrderByID(
	serverURL string, actual, expected *common.ReadResult,
) *testutils.CompareResult {
	sort.Slice(actual.Data, func(i, j int) bool {
		ai, _ := actual.Data[i].Fields["_id"].(string)
		aj, _ := actual.Data[j].Fields["_id"].(string)

		return ai < aj
	})

	return testroutines.ComparatorSubsetRead(serverURL, actual, expected)
}
