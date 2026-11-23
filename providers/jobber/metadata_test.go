package jobber

import (
	"context"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

func introspectionResponse(typeName string) string {
	return `{
		"data": {
			"__type": {
				"name": "` + typeName + `",
				"fields": [
					{"name": "id", "type": {"name": null, "kind": "NON_NULL", "ofType": {"name": "EncodedId", "kind": "SCALAR"}}},
					{"name": "title", "type": {"name": "String", "kind": "SCALAR", "ofType": null}}
				]
			}
		}
	}`
}

func TestListObjectMetadata_RequestsIncludesClientID(t *testing.T) {
	t.Parallel()

	srv := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If:   mockcond.And{mockcond.MethodPOST(), bodyContains(`__type(name: \"Request\")`)},
				Then: mockserver.ResponseString(http.StatusOK, introspectionResponse("Request")),
			},
			{
				If:   mockcond.And{mockcond.MethodPOST(), bodyContains(`__type(name: \"Client\")`)},
				Then: mockserver.ResponseString(http.StatusOK, introspectionResponse("Client")),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
	defer srv.Close()

	conn, err := constructTestConnector(srv.URL)
	assert.NilError(t, err)

	res, err := conn.ListObjectMetadata(context.Background(), []string{objectRequests, objectClients})
	assert.NilError(t, err)
	assert.Equal(t, len(res.Errors), 0)

	requests := res.Result[objectRequests]
	assert.Equal(t, requests.Fields["clientId"].ValueType, common.ValueType(common.ValueTypeString))
	assert.Equal(t, requests.Fields["clientId"].ProviderType, "EncodedId")
	assert.Equal(t, requests.Fields["title"].ValueType, common.ValueType(common.ValueTypeString))

	_, ok := res.Result[objectClients].Fields["clientId"]
	assert.Assert(t, !ok, "clients metadata should not get clientId")
}
