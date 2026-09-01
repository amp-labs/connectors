package acculynx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

// singleRecordServer answers one GET for a record path, asserting the includes
// expansion the connector is expected to send. Only the response body differs
// between the by-id tests, so each case stays a single line.
func singleRecordServer(t *testing.T, path, includes, body string) *httptest.Server {
	t.Helper()

	return mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path(path),
			mockcond.QueryParam("includes", includes),
		},
		Then: mockserver.ResponseString(http.StatusOK, body),
	}.Server()
}

func TestGetRecordsByIds_RejectsUnsupportedObject(t *testing.T) {
	t.Parallel()

	conn, err := constructTestReadConnector("http://unused")
	assert.NilError(t, err)

	_, err = conn.GetRecordsByIds(context.Background(), "leads",
		[]string{"id-1"}, []string{"id"}, nil)

	assert.Assert(t, errors.Is(err, common.ErrGetRecordNotSupportedForObject))
}

func TestGetRecordsByIds_LowercasesObjectName(t *testing.T) {
	t.Parallel()

	conn, err := constructTestReadConnector("http://unused")
	assert.NilError(t, err)

	// "CONTACTS" should be normalized to "contacts" and accepted (empty ids → empty slice, no error).
	rows, err := conn.GetRecordsByIds(context.Background(), "CONTACTS", nil, nil, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(rows), 0)
}

func TestGetRecordsByIds_EmptyIdsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	conn, err := constructTestReadConnector("http://unused")
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectContacts, nil, nil, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(rows), 0)
}

func TestGetRecordsByIds_FetchesEachIdInOrder(t *testing.T) {
	t.Parallel()

	srv := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/jobs/job-1"),
				},
				Then: mockserver.ResponseString(http.StatusOK,
					`{"id":"job-1","jobName":"First","modifiedDate":"2026-05-01T00:00:00Z"}`),
			},
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/jobs/job-2"),
				},
				Then: mockserver.ResponseString(http.StatusOK,
					`{"id":"job-2","jobName":"Second","modifiedDate":"2026-05-02T00:00:00Z"}`),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-1", "job-2"}, []string{"id", "jobName"}, nil)

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 2)

	assert.Equal(t, rows[0].Id, "job-1")
	assert.Equal(t, rows[0].Fields["id"], "job-1")
	assert.Equal(t, rows[0].Fields["jobName"], "First")

	assert.Equal(t, rows[1].Id, "job-2")
	assert.Equal(t, rows[1].Fields["id"], "job-2")
	assert.Equal(t, rows[1].Fields["jobName"], "Second")
}

func TestGetRecordsByIds_RawPreservedUntouched(t *testing.T) {
	t.Parallel()

	body := `{"id":"contact-1","firstName":"Diane","lastName":"Tiblin",` +
		`"_link":"https://api.acculynx.com/api/v2/contacts/contact-1"}`

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/v2/contacts/contact-1"),
		},
		Then: mockserver.ResponseString(http.StatusOK, body),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectContacts,
		[]string{"contact-1"}, []string{"id"}, nil)

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)

	// Fields was filtered to just "id" — but Raw must keep every key the API returned.
	assert.Equal(t, rows[0].Raw["firstName"], "Diane")
	assert.Equal(t, rows[0].Raw["lastName"], "Tiblin")
	assert.Equal(t, rows[0].Raw["_link"], "https://api.acculynx.com/api/v2/contacts/contact-1")
}

func TestGetRecordsByIds_NotFoundIsSkippedNotErrored(t *testing.T) {
	t.Parallel()

	// A 404 for one id must not fail the whole batch: the batch returns the ids
	// that resolve and silently omits the missing one. This is required for the
	// appointment->user edge (company/crew calendar ids are not users and 404),
	// and also hardens hydration for a since-deleted job/contact.
	srv := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/jobs/job-1"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"id":"job-1","jobName":"First"}`),
			},
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/jobs/missing-job"),
				},
				Then: mockserver.ResponseString(http.StatusNotFound, `{"detail":"NotFound"}`),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-1", "missing-job"}, []string{"id", "jobName"}, nil)

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, rows[0].Id, "job-1")
	assert.Equal(t, rows[0].Fields["jobName"], "First")
}

func TestGetRecordsByIds_NonNotFoundStillErrors(t *testing.T) {
	t.Parallel()

	// A non-404 failure (e.g. 500) must still fail the batch — only genuine
	// not-found ids are skipped.
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/v2/jobs/boom-job"),
		},
		Then: mockserver.ResponseString(http.StatusInternalServerError, `{"detail":"boom"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	_, err = conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"boom-job"}, []string{"id"}, nil)

	assert.Assert(t, err != nil, "a 500 from upstream should still surface as an error")
}

func TestGetRecordsByIds_HydratesUsers(t *testing.T) {
	t.Parallel()

	// Users were added to the batch-readable set for the appointment->user edge;
	// a calendar id equals its user id, so hydration resolves the full user
	// (including role, which Hatch filters on).
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/v2/users/794f06f5"),
		},
		Then: mockserver.ResponseString(http.StatusOK,
			`{"id":"794f06f5","displayName":"Diane Hatch","role":{"id":"r1","name":"CompanyAdmin"}}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectUsers,
		[]string{"794f06f5"}, []string{"id", "displayName", "role"}, nil)

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, rows[0].Id, "794f06f5")
	assert.Equal(t, rows[0].Fields["displayName"], "Diane Hatch")
}

func TestGetRecordsByIds_AttachesJobContactsAssociation(t *testing.T) {
	t.Parallel()

	// The enrichment path the server uses for subscription events: a job id plus
	// the association the customer configured. AccuLynx expands contacts inline
	// on the single-record endpoint, so the association is served from the same
	// response — no second request.
	srv := singleRecordServer(t, "/api/v2/jobs/job-001", "initialAppointment,contacts",
		`{"id":"job-001","jobName":"Roof repair","contacts":[
			{"id":"jc-1","isPrimary":true,"contact":{"id":"ctc-100","firstName":"Diane"}}]}`)

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-001"}, []string{"id", "jobName"}, []string{jobContactsAssociation})

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)

	assocs := rows[0].Associations[jobContactsAssociation]
	assert.Equal(t, len(assocs), 1)
	assert.Equal(t, assocs[0].ObjectId, "ctc-100")
	assert.Equal(t, assocs[0].ProviderAssociationMetadata["isPrimary"], true)
}

func TestGetRecordsByIds_NoAssociationsRequestedAttachesNone(t *testing.T) {
	t.Parallel()

	// Without a requested association the job is still fetched with its default
	// initialAppointment expansion, and Associations stays empty.
	srv := singleRecordServer(t, "/api/v2/jobs/job-001", "initialAppointment",
		`{"id":"job-001","contacts":[{"id":"jc-1","contact":{"id":"ctc-100"}}]}`)

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-001"}, []string{"id"}, nil)

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, len(rows[0].Associations), 0)
}

func TestGetRecordsByIds_AttachesEveryContactNotJustPrimary(t *testing.T) {
	t.Parallel()

	// A job commonly carries several contacts. The association must be a list
	// with one entry per contact, each keeping its own isPrimary flag — not just
	// the primary, and not the last one overwriting the rest.
	srv := singleRecordServer(t, "/api/v2/jobs/job-001", "initialAppointment,contacts",
		`{"id":"job-001","contacts":[
			{"id":"jc-1","isPrimary":true,"contact":{"id":"ctc-100","firstName":"Diane"}},
			{"id":"jc-2","isPrimary":false,"contact":{"id":"ctc-200","firstName":"Alan"}},
			{"id":"jc-3","isPrimary":false,"contact":{"id":"ctc-300","firstName":"Marina"}}]}`)

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-001"}, []string{"id"}, []string{jobContactsAssociation})

	assert.NilError(t, err)

	assocs := rows[0].Associations[jobContactsAssociation]
	assert.Equal(t, len(assocs), 3)
	assert.Equal(t, assocs[0].ObjectId, "ctc-100")
	assert.Equal(t, assocs[1].ObjectId, "ctc-200")
	assert.Equal(t, assocs[2].ObjectId, "ctc-300")
	assert.Equal(t, assocs[0].ProviderAssociationMetadata["isPrimary"], true)
	assert.Equal(t, assocs[1].ProviderAssociationMetadata["isPrimary"], false)
}

func TestGetRecordsByIds_EmptyContactsLeavesNoAssociationKey(t *testing.T) {
	t.Parallel()

	// A job with no contacts must not gain an empty "contacts" key — the caller
	// distinguishes "no associations" from "an association with nothing in it".
	srv := singleRecordServer(t, "/api/v2/jobs/job-001", "initialAppointment,contacts",
		`{"id":"job-001","contacts":[]}`)

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-001"}, []string{"id"}, []string{jobContactsAssociation})

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, len(rows[0].Associations), 0)
}

func TestGetRecordsByIds_AssociationsAttachPerRowAcrossBatch(t *testing.T) {
	t.Parallel()

	// Batch fan-out and association extraction have to work together: each row
	// keeps its own contacts, and a job without contacts stays clean even when
	// a sibling in the same batch has them.
	srv := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{mockcond.MethodGET(), mockcond.Path("/api/v2/jobs/job-1")},
				Then: mockserver.ResponseString(http.StatusOK,
					`{"id":"job-1","contacts":[{"id":"jc-1","isPrimary":true,"contact":{"id":"ctc-1"}}]}`),
			},
			{
				If:   mockcond.And{mockcond.MethodGET(), mockcond.Path("/api/v2/jobs/job-2")},
				Then: mockserver.ResponseString(http.StatusOK, `{"id":"job-2"}`),
			},
			{
				If: mockcond.And{mockcond.MethodGET(), mockcond.Path("/api/v2/jobs/job-3")},
				Then: mockserver.ResponseString(http.StatusOK,
					`{"id":"job-3","contacts":[{"id":"jc-3","isPrimary":false,"contact":{"id":"ctc-3"}}]}`),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectJobs,
		[]string{"job-1", "job-2", "job-3"}, []string{"id"}, []string{jobContactsAssociation})

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 3)
	assert.Equal(t, rows[0].Associations[jobContactsAssociation][0].ObjectId, "ctc-1")
	assert.Equal(t, len(rows[1].Associations), 0)
	assert.Equal(t, rows[2].Associations[jobContactsAssociation][0].ObjectId, "ctc-3")
}

func TestGetRecordsByIds_AssociationOnObjectWithoutOneIsIgnored(t *testing.T) {
	t.Parallel()

	// Only jobs define an association today. Asking for one on contacts must be
	// accepted and simply yield nothing, rather than erroring — matching how a
	// read of the same object behaves.
	srv := singleRecordServer(t, "/api/v2/contacts/ctc-100", "emailAddress,phoneNumber",
		`{"id":"ctc-100","firstName":"Diane"}`)

	conn, err := constructTestReadConnector(srv.URL)
	assert.NilError(t, err)

	rows, err := conn.GetRecordsByIds(context.Background(), objectContacts,
		[]string{"ctc-100"}, []string{"id"}, []string{jobContactsAssociation})

	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, len(rows[0].Associations), 0)
}
