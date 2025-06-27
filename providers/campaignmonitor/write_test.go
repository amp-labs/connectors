package campaignmonitor

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	clientsResponse := testutils.DataFromFile(t, "write_clients.json")
	peopleResponse := testutils.DataFromFile(t, "write_people.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the clients",
			Input: common.WriteParams{ObjectName: "clients", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v3.3/clients.json"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, clientsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "aa164e9c8ab0471294fe6148fc9cf634",
				Errors:   nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the people",
			Input: common.WriteParams{ObjectName: "people", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v3.3/clients/744cdce058fc61d9ef5e2492f8d8fbaf/people.json"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, peopleResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Errors:   nil,
				Data: map[string]any{
					"EmailAddress": "sally@sparrow.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the suppress list",
			Input: common.WriteParams{ObjectName: "suppress", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v3.3/clients/744cdce058fc61d9ef5e2492f8d8fbaf/suppress.json"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, nil),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
