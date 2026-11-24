package acculynx

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

// Spec-shaped: one past status and one current. The sandbox has no workflow
// statuses, so a populated live response has not been observed.
//
//go:embed test/read/job-milestone-current-status.json
var jobMilestoneCurrentStatusResponse []byte

// Captured live from the sandbox (ids replaced): an account without workflow
// statuses answers 200 with an empty array.
//
//go:embed test/read/job-milestone-current-empty.json
var jobMilestoneCurrentEmptyResponse []byte

// expectedCurrentWorkflowStatus is the isCurrent entry of
// job-milestone-current-status.json without its server-clock duration.endDate.
func expectedCurrentWorkflowStatus() map[string]any {
	return map[string]any{
		"id": "31081f76-4a5d-4efd-8bc3-22a1a3fbb83f",
		"_link": "https://api.acculynx.com/api/v2/jobs/job-001/milestones/" +
			"893c971d-4dba-4b95-ae1e-a4ba477b906c/statuses/31081f76-4a5d-4efd-8bc3-22a1a3fbb83f",
		"name":      "Follow Up",
		"duration":  map[string]any{"startDate": "2023-11-10T15:00:00Z"},
		"isCurrent": true,
		"workflowStatus": map[string]any{
			"id":   "40be7484-ece1-4a01-9b4c-2de030d0c841",
			"name": "Follow Up",
		},
	}
}

// workflowStatusServer serves job-001 and job-002 from the jobs list and by id,
// empty custom-field definitions, and each job's milestones/current from
// statusByJob, counting the status calls so tests can assert none were made.
func workflowStatusServer(
	statusByJob map[string]http.HandlerFunc,
	statusCalls *atomic.Int32,
) *httptest.Server {
	cases := []mockserver.Case{
		{
			If: mockcond.And{
				mockcond.MethodGET(),
				mockcond.Path("/api/v2/jobs"),
				mockcond.QueryParam("includes", "initialAppointment"),
			},
			Then: mockserver.Response(http.StatusOK, jobsWithContactsResponse),
		},
		{
			If: mockcond.And{
				mockcond.MethodGET(),
				mockcond.Path("/api/v2/company-settings/custom-fields"),
			},
			Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
		},
	}

	for _, jobID := range []string{"job-001", "job-002"} {
		status := statusByJob[jobID]

		cases = append(cases,
			mockserver.Case{
				If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/api/v2/jobs/" + jobID)},
				Then: mockserver.ResponseString(http.StatusOK, `{"id":"`+jobID+`"}`),
			},
			mockserver.Case{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/jobs/" + jobID + "/milestones/current"),
					mockcond.QueryParam("includes", "status"),
				},
				Then: func(w http.ResponseWriter, r *http.Request) {
					statusCalls.Add(1)
					status(w, r)
				},
			},
		)
	}

	return mockserver.Switch{
		Setup:   mockserver.ContentJSON(),
		Cases:   cases,
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
}

type workflowStatusCase struct {
	name        string
	fields      []string
	statusByJob map[string]http.HandlerFunc
	wantCalls   int32
	// want maps job id to the expected currentWorkflowStatus; a job missing
	// from the map must not carry the field at all.
	want map[string]any
}

func workflowStatusCases() []workflowStatusCase {
	found := mockserver.Response(http.StatusOK, jobMilestoneCurrentStatusResponse)
	empty := mockserver.Response(http.StatusOK, jobMilestoneCurrentEmptyResponse)
	notFound := mockserver.ResponseString(http.StatusNotFound, `{"title":"Unable to find the requested job","status":404}`)

	return []workflowStatusCase{
		{
			name:        "selected: the current entry, or null for a job without one",
			fields:      []string{"id", "currentWorkflowStatus"},
			statusByJob: map[string]http.HandlerFunc{"job-001": found, "job-002": empty},
			wantCalls:   2,
			want:        map[string]any{"job-001": expectedCurrentWorkflowStatus(), "job-002": nil},
		},
		{
			name:        "selected in lowercase",
			fields:      []string{"id", "currentworkflowstatus"},
			statusByJob: map[string]http.HandlerFunc{"job-001": found, "job-002": empty},
			wantCalls:   2,
			want:        map[string]any{"job-001": expectedCurrentWorkflowStatus(), "job-002": nil},
		},
		{
			name:        "not selected: no status calls and no field",
			fields:      []string{"id", "currentMilestone"},
			statusByJob: map[string]http.HandlerFunc{"job-001": found, "job-002": found},
			wantCalls:   0,
			want:        map[string]any{},
		},
		{
			name:        "404 leaves that job without the field and the rest succeeds",
			fields:      []string{"id", "currentWorkflowStatus"},
			statusByJob: map[string]http.HandlerFunc{"job-001": notFound, "job-002": found},
			wantCalls:   2,
			want:        map[string]any{"job-002": expectedCurrentWorkflowStatus()},
		},
	}
}

// assertWorkflowStatus checks the field in both Fields (under fieldsKey) and Raw.
func assertWorkflowStatus(t *testing.T, row common.ReadResultRow, fieldsKey string, want map[string]any) {
	t.Helper()

	wantValue, wantSet := want[row.Id]

	gotFields, fieldsSet := row.Fields[fieldsKey]
	assert.Equal(t, fieldsSet, wantSet, "Fields presence for %s", row.Id)
	assert.DeepEqual(t, gotFields, wantValue)

	gotRaw, rawSet := row.Raw[fieldCurrentWorkflowStatus]
	assert.Equal(t, rawSet, wantSet, "Raw presence for %s", row.Id)
	assert.DeepEqual(t, gotRaw, wantValue)
}

func TestReadJobsCurrentWorkflowStatus(t *testing.T) {
	t.Parallel()

	for _, tt := range workflowStatusCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32

			srv := workflowStatusServer(tt.statusByJob, &calls)
			defer srv.Close()

			conn, err := constructTestReadConnector(srv.URL)
			assert.NilError(t, err)

			res, err := conn.Read(context.Background(), common.ReadParams{
				ObjectName: objectJobs,
				Fields:     connectors.Fields(tt.fields...),
			})
			assert.NilError(t, err)
			assert.Equal(t, len(res.Data), 2)
			assert.Equal(t, calls.Load(), tt.wantCalls)

			for _, row := range res.Data {
				assertWorkflowStatus(t, row, "currentworkflowstatus", tt.want)
			}
		})
	}
}

func TestGetRecordsByIdsJobsCurrentWorkflowStatus(t *testing.T) {
	t.Parallel()

	for _, tt := range workflowStatusCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32

			srv := workflowStatusServer(tt.statusByJob, &calls)
			defer srv.Close()

			conn, err := constructTestReadConnector(srv.URL)
			assert.NilError(t, err)

			res, err := conn.GetRecordsByIds(context.Background(), objectJobs,
				[]string{"job-001", "job-002"}, tt.fields, nil)
			assert.NilError(t, err)
			assert.Equal(t, len(res.Rows), 2)
			assert.Equal(t, calls.Load(), tt.wantCalls)

			// pickFields keeps the caller's spelling of the requested name.
			for _, row := range res.Rows {
				assertWorkflowStatus(t, row, tt.fields[1], tt.want)
			}
		})
	}
}
