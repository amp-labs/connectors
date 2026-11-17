package acculynx

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

// jobs/financials and jobs/representatives rows carry the parent job from the
// /jobs list page as the Job association's Raw, so the server skips its
// GET /jobs/{id} per row. The list must ask for the same includes as that GET.
func TestReadJobAssociationCarriesListJob(t *testing.T) {
	t.Parallel()

	const jobID = "875b28e8-4f10-44e2-9af1-aff90527291e"

	var list struct {
		Items []map[string]any `json:"items"`
	}

	assert.NilError(t, json.Unmarshal(jobsWithInitialAppointmentResponse, &list))

	listJob := list.Items[0]

	tests := []struct {
		object string
		leaf   []byte
		rows   int
	}{
		{object: "jobs/financials", leaf: jobFinancials001Response, rows: 1},
		// Two representatives on one job: each row gets the same job.
		{object: "jobs/representatives", leaf: jobRepresentativesResponse, rows: 2},
	}

	for _, tt := range tests {
		t.Run(tt.object, func(t *testing.T) {
			t.Parallel()

			server := mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("includes", "initialAppointment"),
						},
						Then: mockserver.Response(http.StatusOK, jobsWithInitialAppointmentResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/" + jobsLeafPath(jobID, tt.object)),
						},
						Then: mockserver.Response(http.StatusOK, tt.leaf),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server()
			defer server.Close()

			conn, err := constructTestConnector(server.URL)
			assert.NilError(t, err)

			result, err := conn.Read(t.Context(), common.ReadParams{
				ObjectName:        tt.object,
				Fields:            connectors.Fields("id"),
				AssociatedObjects: []string{"jobs"},
			})
			assert.NilError(t, err)
			assert.Equal(t, len(result.Data), tt.rows)

			for _, row := range result.Data {
				assert.DeepEqual(t, row.Associations["jobs"], []common.Association{{ObjectId: jobID, Raw: listJob}})
			}
		})
	}
}

// jobsLeafPath turns "jobs/financials" into "jobs/{jobID}/financials".
func jobsLeafPath(jobID, object string) string {
	return "jobs/" + jobID + object[len("jobs"):]
}

// The job's embedded contacts only have phone/email stubs, so jobs -> contacts
// keeps an empty Raw and the server still fetches each contact.
func TestReadJobContactsKeepEmptyRaw(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/jobs"),
					mockcond.QueryParam("includes", "initialAppointment,contacts"),
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
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	assert.NilError(t, err)

	result, err := conn.Read(t.Context(), common.ReadParams{
		ObjectName:        "jobs",
		Fields:            connectors.Fields("id"),
		AssociatedObjects: []string{"contacts"},
	})
	assert.NilError(t, err)
	assert.Equal(t, len(result.Data), 2)

	for _, row := range result.Data {
		assocs := row.Associations["contacts"]
		assert.Assert(t, len(assocs) > 0)

		for _, assoc := range assocs {
			assert.Assert(t, assoc.Raw == nil, "contact %s has Raw", assoc.ObjectId)
		}
	}
}

func TestAttachFinancialAssociations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		params   common.ReadParams
		parentID string
		raw      map[string]any
		expected map[string][]common.Association
	}{
		{
			name:     "uses the fan-out parent job id",
			params:   common.ReadParams{ObjectName: "jobs/financials", AssociatedObjects: []string{"jobs"}},
			parentID: "job-parent",
			raw:      map[string]any{"jobId": "job-001"},
			expected: map[string][]common.Association{"jobs": {{ObjectId: "job-parent"}}},
		},
		{
			name:     "attaches nothing without a parent job id",
			params:   common.ReadParams{ObjectName: "jobs/financials", AssociatedObjects: []string{"jobs"}},
			raw:      map[string]any{"jobId": "job-001"},
			expected: nil,
		},
		{
			name:     "attaches nothing when the association is not requested",
			params:   common.ReadParams{ObjectName: "jobs/financials"},
			parentID: "job-parent",
			raw:      map[string]any{"jobId": "job-001"},
			expected: nil,
		},
		{
			name:     "ignores other objects",
			params:   common.ReadParams{ObjectName: "jobs/invoices", AssociatedObjects: []string{"jobs"}},
			parentID: "job-parent",
			raw:      map[string]any{"jobId": "job-001"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rows := []common.ReadResultRow{{Raw: tt.raw}}

			attachFinancialAssociations(tt.params, tt.parentID, nil, rows)

			assert.DeepEqual(t, rows[0].Associations, tt.expected)
		})
	}
}
