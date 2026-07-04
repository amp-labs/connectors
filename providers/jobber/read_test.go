package jobber

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}

// bodyContains matches requests whose raw body contains the substring.
// GraphQL queries arrive JSON-encoded, so quotes appear escaped (\").
func bodyContains(substring string) mockcond.Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return false
		}

		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		return strings.Contains(string(body), substring)
	}
}

func TestRead_IncrementalClientsUsesUpdatedAtFilter(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains(`updatedAt:`),
			bodyContains(`after: \"2026-07-01T00:00:00Z\"`),
			bodyContains(`before: \"2026-07-03T00:00:00Z\"`),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"clients": {
					"nodes": [
						{"id": "C1", "name": "Ada", "updatedAt": "2026-07-02T10:00:00Z"}
					],
					"pageInfo": {"endCursor": "CUR1", "hasNextPage": true},
					"totalCount": 1
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: "clients",
		Fields:     connectors.Fields("id", "name"),
		Since:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Until:      time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC),
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(1))
	assert.Equal(t, res.NextPage.String(), "CUR1")
	assert.Equal(t, res.Data[0].Fields["name"], "Ada")
}

func TestRead_IncrementalVisitsUsesCreatedAtFilter(t *testing.T) {
	t.Parallel()

	// Visits have no updatedAt in Jobber, so incremental reads filter on
	// createdAt and never capture updates to existing records.
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains(`createdAt:`),
			bodyContains(`after: \"2026-07-01T00:00:00Z\"`),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"visits": {
					"nodes": [
						{"id": "V1", "title": "Mow lawn", "createdAt": "2026-07-02T08:00:00Z"}
					],
					"pageInfo": {"endCursor": "CURV", "hasNextPage": false},
					"totalCount": 1
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: "visits",
		Fields:     connectors.Fields("id", "title"),
		Since:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(1))
	assert.Equal(t, res.NextPage.String(), "")
}

func TestRead_IncrementalJobsSortsAndCutsOffOldRecords(t *testing.T) {
	t.Parallel()

	// Jobs cannot be filtered by updatedAt, so the connector sorts by
	// UPDATED_AT descending and cuts off client-side. The third record is
	// older than Since: it must be dropped and pagination must stop even
	// though the server reports another page.
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.MethodPOST(),
			bodyContains("UPDATED_AT"),
			bodyContains("DESCENDING"),
		},
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"jobs": {
					"nodes": [
						{"id": "J3", "title": "newest", "updatedAt": "2026-07-03T10:00:00Z"},
						{"id": "J2", "title": "recent", "updatedAt": "2026-07-02T10:00:00Z"},
						{"id": "J1", "title": "stale", "updatedAt": "2026-06-30T10:00:00Z"}
					],
					"pageInfo": {"endCursor": "CURJ", "hasNextPage": true},
					"totalCount": 3
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: objectJobs,
		Fields:     connectors.Fields("id", "title"),
		Since:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(2))
	assert.Equal(t, res.Data[0].Id, "J3")
	assert.Equal(t, res.Data[1].Id, "J2")
	assert.Equal(t, res.NextPage.String(), "")
	assert.Equal(t, res.Done, true)
}

func TestRead_IncrementalJobsKeepsPaginatingWhileRecordsAreNew(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"jobs": {
					"nodes": [
						{"id": "J3", "updatedAt": "2026-07-03T10:00:00Z"},
						{"id": "J2", "updatedAt": "2026-07-02T10:00:00Z"}
					],
					"pageInfo": {"endCursor": "CURJ", "hasNextPage": true},
					"totalCount": 2
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: objectJobs,
		Fields:     connectors.Fields("id"),
		Since:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(2))
	assert.Equal(t, res.NextPage.String(), "CURJ")
}

func TestRead_IncrementalJobsDropsRecordsNewerThanUntil(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"jobs": {
					"nodes": [
						{"id": "J3", "updatedAt": "2026-07-04T10:00:00Z"},
						{"id": "J2", "updatedAt": "2026-07-02T10:00:00Z"}
					],
					"pageInfo": {"endCursor": "CURJ", "hasNextPage": true},
					"totalCount": 2
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: objectJobs,
		Fields:     connectors.Fields("id"),
		Since:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Until:      time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC),
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(1))
	assert.Equal(t, res.Data[0].Id, "J2")
	assert.Equal(t, res.NextPage.String(), "CURJ")
}

func TestRead_FullJobsReadDoesNotSort(t *testing.T) {
	t.Parallel()

	// Without Since the query must not include the UPDATED_AT sort, keeping
	// full reads on the API's default ordering.
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    bodyContains("UPDATED_AT"),
		Then:  mockserver.ResponseString(http.StatusBadRequest, `{"errors":[{"message":"unexpected sort"}]}`),
		Else: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"jobs": {
					"nodes": [
						{"id": "J1", "updatedAt": "2026-06-30T10:00:00Z"}
					],
					"pageInfo": {"endCursor": "CURJ", "hasNextPage": true},
					"totalCount": 1
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: objectJobs,
		Fields:     connectors.Fields("id"),
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(1))
	assert.Equal(t, res.NextPage.String(), "CURJ")
}

func TestRead_SurfacesGraphQLErrors(t *testing.T) {
	t.Parallel()

	// GraphQL failures such as throttling arrive with HTTP 200, an "errors"
	// array and no "data"; the response interceptor rewrites the status code
	// so the error interpreter surfaces the API's message as a typed error.
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then: mockserver.ResponseString(http.StatusOK, `{
			"errors": [{"message": "Throttled", "extensions": {"code": "THROTTLED",
				"documentation": "https://developer.getjobber.com/docs/using_jobbers_api/api_rate_limits"}}]
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	_, err = conn.Read(context.Background(), common.ReadParams{
		ObjectName: objectJobs,
		Fields:     connectors.Fields("id"),
		Since:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.ErrorContains(t, err, "Throttled")
	assert.Assert(t, errors.Is(err, common.ErrLimitExceeded))
}

func TestRead_SurfacesGraphQLAuthErrors(t *testing.T) {
	t.Parallel()

	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    mockcond.MethodPOST(),
		Then: mockserver.ResponseString(http.StatusOK, `{
			"errors": [{"message": "Auth required", "extensions": {"code": "UNAUTHENTICATED"}}],
			"data": null
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	_, err = conn.Read(context.Background(), common.ReadParams{
		ObjectName: "clients",
		Fields:     connectors.Fields("id"),
	})

	assert.ErrorContains(t, err, "Auth required")
	assert.Assert(t, errors.Is(err, common.ErrAccessToken))
}

func TestRead_NextPageCursorIsQuotedInQuery(t *testing.T) {
	t.Parallel()

	// Cursors are base64 and must be sent as GraphQL strings; unquoted they
	// fail to parse server-side.
	srv := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    bodyContains(`after: \"eyJsYXN0X2lkIjo0fQ==\"`),
		Then: mockserver.ResponseString(http.StatusOK, `{
			"data": {
				"clients": {
					"nodes": [
						{"id": "C2", "name": "Grace", "updatedAt": "2026-07-02T10:00:00Z"}
					],
					"pageInfo": {"endCursor": "", "hasNextPage": false},
					"totalCount": 1
				}
			}
		}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.Read(context.Background(), common.ReadParams{
		ObjectName: "clients",
		Fields:     connectors.Fields("id", "name"),
		NextPage:   "eyJsYXN0X2lkIjo0fQ==",
	})

	assert.NilError(t, err)
	assert.Equal(t, res.Rows, int64(1))
	assert.Equal(t, res.Data[0].Fields["name"], "Grace")
}
