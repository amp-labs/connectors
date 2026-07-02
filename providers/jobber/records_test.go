package jobber

import (
	"context"
	"errors"
	"net/http"
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

	rows, err := conn.GetRecordsByIds(context.Background(), "clients", nil, nil, nil)
	assert.NilError(t, err)
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
	rows, err := conn.GetRecordsByIds(context.Background(),
		"clients", []string{"Z2lkOi8vSm9iYmVyL0NsaWVudC8x"}, []string{"firstname"}, nil)
	assert.NilError(t, err)
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

	_, err = conn.GetRecordsByIds(context.Background(), "clients", []string{"MQ=="}, nil, nil)
	assert.Assert(t, errors.Is(err, errRecordFetchNotFound))
}

func TestGetRecordsByIds_GraphQLErrorSurfaces(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then: mockserver.ResponseString(http.StatusOK,
			`{"errors": [{"message": "Visit not found"}], "data": null}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	_, err = conn.GetRecordsByIds(context.Background(), "visits", []string{"MQ=="}, nil, nil)
	assert.Assert(t, errors.Is(err, errRecordFetchFailed))
}
