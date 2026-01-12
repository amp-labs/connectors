package cloudtalk

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name: "Delete Contact",
			Input: common.DeleteParams{
				ObjectName: "contacts",
				RecordId:   "123",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/contacts/delete/123.json"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{}`)),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
		},
		{
			Name: "Delete Tag",
			Input: common.DeleteParams{
				ObjectName: "tags",
				RecordId:   "555",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/tags/delete/555.json"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{}`)),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
