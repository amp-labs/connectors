package jobber

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

// Every batch-readable object must resolve to an embedded singular getter
// query file; this guards the singularization + file-name mapping.
func TestBatchReadableObjectsHaveGetterQueries(t *testing.T) {
	t.Parallel()

	for _, objectName := range batchReadableObjects.List() {
		_, err := graphql.Operation(queryFiles, "query", getObjectName(objectName), nil)
		assert.NilError(t, err, "missing singular getter query for object %s", objectName)
	}
}

func TestGetRecordsByIds_UnsupportedObject(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	_, err := conn.GetRecordsByIds(context.Background(), "vehicles", []string{"MQ=="}, nil, nil)
	assert.Assert(t, errors.Is(err, common.ErrGetRecordNotSupportedForObject))
}

func TestGetRecordsByIds_EmptyIds(t *testing.T) {
	t.Parallel()

	conn := &Connector{}

	batchRes, err := conn.GetRecordsByIds(context.Background(), "clients", nil, nil, nil)
	assert.NilError(t, err)

	rows := batchRes.Rows
	assert.Equal(t, len(rows), 0)
}

func TestGetRecordsByIds_FetchesAndFiltersFields(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains(`"id"`),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"client": {
					"id": "Z2lkOi8vSm9iYmVyL0NsaWVudC8x",
					"firstName": "Webhook",
					"lastName": "SpikeTest",
					"isLead": true
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	// The framework lower-cases requested field names; matching must be
	// case-insensitive against Jobber's camelCase keys.
	batchRes, err := conn.GetRecordsByIds(context.Background(),
		"clients", []string{"Z2lkOi8vSm9iYmVyL0NsaWVudC8x"}, []string{"firstname"}, nil)
	assert.NilError(t, err)

	rows := batchRes.Rows
	assert.Equal(t, len(rows), 1)

	assert.Equal(t, rows[0].Id, "Z2lkOi8vSm9iYmVyL0NsaWVudC8x")
	assert.Equal(t, rows[0].Fields["firstname"], "Webhook")
	assert.Equal(t, len(rows[0].Fields), 1)
	assert.Equal(t, rows[0].Raw["lastName"], "SpikeTest")
}

func TestGetRecordsByIds_RecordNotFound(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then:  mockserver.ResponseString(http.StatusOK, `{"data": {"client": null}}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	batchRes, err := conn.GetRecordsByIds(context.Background(), "clients", []string{"MQ=="}, nil, nil)

	// A missing record is reported as a not-found failure for its id.
	assert.NilError(t, err)
	assert.Equal(t, len(batchRes.Rows), 0)
	assert.Equal(t, len(batchRes.Failures), 1)
	assert.Equal(t, batchRes.Failures[0].RecordId, "MQ==")
	assert.Equal(t, batchRes.Failures[0].Reason, common.FailureReasonNotFound)
	assert.Assert(t, errors.Is(batchRes.Failures[0].Err, errRecordFetchNotFound))
}

func TestGetRecordsByIds_GraphQLErrorSurfaces(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then: mockserver.ResponseString(http.StatusOK,
			`{"errors": [{"message": "Internal server error"}], "data": null}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	batchRes, err := conn.GetRecordsByIds(context.Background(), "visits", []string{"MQ=="}, nil, nil)

	// A GraphQL error is reported as a failure for its id and returned as the error.
	assert.ErrorIs(t, err, errRecordFetchFailed)
	assert.ErrorContains(t, err, "record MQ==")
	assert.Equal(t, len(batchRes.Rows), 0)
	assert.Equal(t, len(batchRes.Failures), 1)
	assert.Equal(t, batchRes.Failures[0].RecordId, "MQ==")
	assert.Assert(t, errors.Is(batchRes.Failures[0].Err, errRecordFetchFailed))
}

// Jobber's throttling response: an "errors" entry with code THROTTLED and no "data".
// https://developer.getjobber.com/docs/using_jobbers_api/api_rate_limits
const throttledResponse = `{
	"errors": [{"message": "Throttled", "extensions": {"code": "THROTTLED"}}],
	"extensions": {"cost": {"requestedQueryCost": 142, "actualQueryCost": 0}}
}`

func getClientsByIds(t *testing.T, server *mockserver.Switch, ids ...string) (*common.BatchReadResult, error) {
	t.Helper()

	srv := server.Server()
	t.Cleanup(srv.Close)

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	return conn.GetRecordsByIds(t.Context(), "clients", ids, nil, nil)
}

func respondAlways(body string) *mockserver.Switch {
	return &mockserver.Switch{
		Setup:   mockserver.ContentJSON(),
		Default: mockserver.ResponseString(http.StatusOK, body),
	}
}

// A THROTTLED error is reported as rate-limited and returned as the error.
func TestGetRecordsByIds_ThrottledIsRateLimited(t *testing.T) {
	t.Parallel()

	batchRes, err := getClientsByIds(t, respondAlways(throttledResponse), "MQ==")

	assert.ErrorIs(t, err, common.ErrLimitExceeded)
	assert.ErrorIs(t, err, errRecordFetchFailed)
	assert.ErrorContains(t, err, "Throttled")
	assert.Equal(t, len(batchRes.Rows), 0)
	assert.Equal(t, len(batchRes.Failures), 1)
	assert.Equal(t, batchRes.Failures[0].RecordId, "MQ==")
	assert.Equal(t, batchRes.Failures[0].Reason, common.FailureReasonRateLimited)
}

// An UNAUTHENTICATED error is reported as an authentication failure, even when the record node is null.
func TestGetRecordsByIds_UnauthenticatedIsAuthFailure(t *testing.T) {
	t.Parallel()

	batchRes, err := getClientsByIds(t, respondAlways(`{
		"errors": [{"message": "Unauthenticated", "extensions": {"code": "UNAUTHENTICATED"}}],
		"data": {"client": null}
	}`), "MQ==")

	assert.ErrorIs(t, err, common.ErrAccessToken)
	assert.ErrorIs(t, err, errRecordFetchFailed)
	assert.Equal(t, len(batchRes.Failures), 1)
	assert.Equal(t, batchRes.Failures[0].Reason, common.FailureReasonPermanent)
}

// An error alongside a null record node is reported as not found and is not returned as the error.
func TestGetRecordsByIds_ErrorWithNullRecordIsNotFound(t *testing.T) {
	t.Parallel()

	batchRes, err := getClientsByIds(t, respondAlways(`{
		"errors": [{"message": "Record not found", "path": ["client"]}],
		"data": {"client": null}
	}`), "MQ==")

	assert.NilError(t, err)
	assert.Equal(t, len(batchRes.Rows), 0)
	assert.Equal(t, len(batchRes.Failures), 1)
	assert.Equal(t, batchRes.Failures[0].RecordId, "MQ==")
	assert.Equal(t, batchRes.Failures[0].Reason, common.FailureReasonNotFound)
	assert.Assert(t, errors.Is(batchRes.Failures[0].Err, errRecordFetchNotFound))
	assert.ErrorContains(t, batchRes.Failures[0].Err, "Record not found")
}

// An error alongside a returned record stays a fetch failure, as before.
func TestGetRecordsByIds_ErrorWithRecordIsFetchFailure(t *testing.T) {
	t.Parallel()

	batchRes, err := getClientsByIds(t, respondAlways(`{
		"errors": [{"message": "Field failed", "path": ["client", "balance"]}],
		"data": {"client": {"id": "MQ==", "balance": null}}
	}`), "MQ==")

	assert.ErrorIs(t, err, errRecordFetchFailed)
	assert.ErrorContains(t, err, "Field failed")
	assert.Equal(t, len(batchRes.Rows), 0)
	assert.Equal(t, len(batchRes.Failures), 1)
	assert.Assert(t, !errors.Is(batchRes.Failures[0].Err, common.ErrNotFound))
}

// Rows of the ids that resolved are returned alongside the failures of the others.
func TestGetRecordsByIds_PartialSuccessKeepsRows(t *testing.T) {
	t.Parallel()

	server := &mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If:   bodyContains(`"id":"found"`),
			Then: mockserver.ResponseString(http.StatusOK, `{"data": {"client": {"id": "found", "firstName": "A"}}}`),
		}, {
			If: bodyContains(`"id":"missing"`),
			Then: mockserver.ResponseString(http.StatusOK,
				`{"errors": [{"message": "Record not found"}], "data": {"client": null}}`),
		}, {
			If:   bodyContains(`"id":"throttled"`),
			Then: mockserver.ResponseString(http.StatusOK, throttledResponse),
		}},
	}

	batchRes, err := getClientsByIds(t, server, "found", "missing", "throttled")

	assert.ErrorIs(t, err, common.ErrLimitExceeded)
	assert.ErrorContains(t, err, "record throttled")
	assert.Assert(t, !strings.Contains(err.Error(), "record missing"))
	assert.Equal(t, len(batchRes.Rows), 1)
	assert.Equal(t, batchRes.Rows[0].Id, "found")
	assert.Equal(t, len(batchRes.Failures), 2)
	assert.Equal(t, batchRes.Failures[0].RecordId, "missing")
	assert.Equal(t, batchRes.Failures[0].Reason, common.FailureReasonNotFound)
	assert.Equal(t, batchRes.Failures[1].RecordId, "throttled")
	assert.Equal(t, batchRes.Failures[1].Reason, common.FailureReasonRateLimited)
}
