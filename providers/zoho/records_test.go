package zoho

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")

	tests := []testconn.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name: "Empty record IDs returns error",
			Input: GetRecordsByIdsInput{
				ObjectName: "contacts",
				Fields:     []string{"Full_Name"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Successfully fetch contacts by IDs",
			Input: GetRecordsByIdsInput{
				ObjectName: "contacts",
				Fields:     []string{"Full_Name"},
				Ids:        []string{"6493490000001291001", "6493490000001291002"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v6/contacts"),
					mockcond.QueryParam("ids", "6493490000001291001,6493490000001291002"),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "6493490000001291001",
					Fields: map[string]any{
						"id":        "6493490000001291001",
						"full_name": "Ryan Dahl2",
					},
					Raw: map[string]any{
						"Full_Name":    "Ryan Dahl2",
						"id":           "6493490000001291001",
						"Assistant":    nil,
						"Created_Time": "2024-12-20T10:09:52+03:00",
						"Created_By": map[string]any{
							"name":  "Joseph Karage",
							"id":    "6493490000000486001",
							"email": "josephkarage@gmail.com",
						},
					},
				},
				{
					Id: "6493490000001291002",
					Fields: map[string]any{
						"id":        "6493490000001291002",
						"full_name": "Jane Smith",
					},
					Raw: map[string]any{
						"Full_Name": "Jane Smith",
						"id":        "6493490000001291002",
						"Created_By": map[string]any{
							"name":  "Joseph Karage",
							"id":    "6493490000000486001",
							"email": "josephkarage@gmail.com",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, tt.Input.Associations)

			var rows []common.ReadResultRow
			if result != nil {
				rows = result.Rows
			}

			tt.Validate(t, err, rows)

			// Every requested id is in the response, so none may be reported missing.
			if err == nil {
				testconn.RequireNoFailures(t, result)
			}
		})
	}
}

// Ids the batch response leaves out are reported as not-found failures, in requested order.
func TestGetRecordsByIds_ReportsMissingIds(t *testing.T) {
	t.Parallel()

	response := testutils.DataFromFile(t, "contacts.json")

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/crm/v6/contacts"),
		Then:  mockserver.Response(http.StatusOK, response),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	// The response holds "6493490000001291001", "6493490000001291002"; the other ids are absent from it.
	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"6493490000001291098", "6493490000001291001", "6493490000001291099", "6493490000001291002"}, []string{"Full_Name"}, nil)
	if err != nil {
		t.Fatalf("a partial batch response must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result, "6493490000001291001", "6493490000001291002")
	testconn.RequireMissingIds(t, result, "6493490000001291098", "6493490000001291099")
}

// makeIds returns count ids starting at first, as strings.
func makeIds(first, count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = strconv.Itoa(first + i)
	}

	return ids
}

// A 204 with no body returns no rows, so every requested id is reported as not found and the call
// does not fail.
func TestGetRecordsByIds_NoContentReportsEveryId(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/crm/v6/contacts"),
		Then:  mockserver.ResponseString(http.StatusNoContent, ""),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"6493490000001291098", "6493490000001291099"}, []string{"Full_Name"}, nil)
	if err != nil {
		t.Fatalf("a 204 must not fail the call: %v", err)
	}

	testconn.RequireRowIds(t, result)
	testconn.RequireMissingIds(t, result, "6493490000001291098", "6493490000001291099")
}

// Zoho documents 404 as "Invalid request", so it fails the call as a bad request and is not read as
// the records being missing.
func TestGetRecordsByIds_NotFoundIsBadRequest(t *testing.T) {
	t.Parallel()

	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.Path("/crm/v6/contacts"),
		Then: mockserver.ResponseString(http.StatusNotFound,
			`{"status":"error"}`),
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts",
		[]string{"6493490000001291001"}, []string{"Full_Name"}, nil)

	if !errors.Is(err, common.ErrBadRequest) {
		t.Errorf("expected common.ErrBadRequest, got %v", err)
	}

	if errors.Is(err, common.ErrNotFound) {
		t.Errorf("error must not wrap common.ErrNotFound: %v", err)
	}

	if httpErr, ok := errors.AsType[*common.HTTPError](err); ok {
		t.Errorf("error must not carry the provider's HTTP error, got status %d", httpErr.Status)
	}

	if result != nil {
		t.Errorf("expected no result, got %+v", result)
	}
}

// More ids than fit on one page are requested in groups of 200 with per_page=200, and the rows of
// every group are returned.
func TestGetRecordsByIds_SplitsIdsIntoPages(t *testing.T) {
	t.Parallel()

	ids := makeIds(1001, 250)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.And{
				mockcond.Path("/crm/v6/contacts"),
				mockcond.QueryParam("per_page", "200"),
				mockcond.QueryParam("ids", strings.Join(ids[:200], ",")),
			},
			Then: mockserver.ResponseString(http.StatusOK,
				`{"data":[{"id":"1001"},{"id":"1200"}],"info":{"more_records":false}}`),
		}, {
			If: mockcond.And{
				mockcond.Path("/crm/v6/contacts"),
				mockcond.QueryParam("per_page", "200"),
				mockcond.QueryParam("ids", strings.Join(ids[200:], ",")),
			},
			Then: mockserver.ResponseString(http.StatusOK,
				`{"data":[{"id":"1201"},{"id":"1250"}],"info":{"more_records":false}}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", ids, []string{"id"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	returned := []string{"1001", "1200", "1201", "1250"}
	testconn.RequireRowIds(t, result, returned...)

	// Every id that neither response returned is reported as not found.
	missing := slices.DeleteFunc(slices.Clone(ids), func(id string) bool {
		return slices.Contains(returned, id)
	})
	testconn.RequireMissingIds(t, result, missing...)
}

// When one group of ids fails, its ids are reported as failures and returned in the error, while the
// rows of the other group are kept.
func TestGetRecordsByIds_FailedPageKeepsOtherRows(t *testing.T) {
	t.Parallel()

	ids := makeIds(1001, 250)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.And{
				mockcond.Path("/crm/v6/contacts"),
				mockcond.QueryParam("ids", strings.Join(ids[:200], ",")),
			},
			Then: mockserver.ResponseString(http.StatusInternalServerError,
				`{"status":"error"}`),
		}, {
			If: mockcond.And{
				mockcond.Path("/crm/v6/contacts"),
				mockcond.QueryParam("ids", strings.Join(ids[200:], ",")),
			},
			Then: mockserver.ResponseString(http.StatusOK,
				`{"data":[{"id":"1201"},{"id":"1250"}],"info":{"more_records":false}}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	result, err := conn.GetRecordsByIds(t.Context(), "contacts", ids, []string{"id"}, nil)
	if !errors.Is(err, common.ErrServer) {
		t.Fatalf("expected the failed request's error, got %v", err)
	}

	testconn.RequireRowIds(t, result, "1201", "1250")

	var failedIds, notFoundIds []string

	for _, failure := range result.Failures {
		if failure.Reason == common.FailureReasonNotFound {
			notFoundIds = append(notFoundIds, failure.RecordId)
		} else {
			failedIds = append(failedIds, failure.RecordId)
		}
	}

	if !slices.Equal(failedIds, ids[:200]) {
		t.Errorf("failed ids %v, want the first 200 requested ids", failedIds)
	}

	// The second group returned 1201 and 1250; its other ids are not found.
	wantNotFound := slices.DeleteFunc(slices.Clone(ids[200:]), func(id string) bool {
		return id == "1201" || id == "1250"
	})
	if !slices.Equal(notFoundIds, wantNotFound) {
		t.Errorf("not-found ids %v, want %v", notFoundIds, wantNotFound)
	}
}
