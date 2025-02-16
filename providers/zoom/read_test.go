package zoom

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestReadModuleUser(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUsers := testutils.DataFromFile(t, "read-users.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "tiger", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Read list of all users",
			Input: common.ReadParams{
				ObjectName: "users", Fields: connectors.Fields("id", "email"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    "KDcuGIm1QgePTO8WbOqwIQ",
						"email": "jchill@example.com",
					},
					Raw: map[string]any{
						"id":              "KDcuGIm1QgePTO8WbOqwIQ",
						"email":           "jchill@example.com",
						"user_created_at": "2019-06-01T07:58:03Z",
						"status":          "active",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/users?next_page_token=8V8HigQkzm2O5r9RUn31D9ZyJHgrmFfbLa2&page_size=300", //nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleUser)
			})
		})
	}
}
