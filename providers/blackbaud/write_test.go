package blackbaud

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the Crm Administration batches",
			Input: common.WriteParams{ObjectName: "crm-adnmg/batches", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm-adnmg/batches"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(
					http.StatusOK, []byte(`{"id": "526b4319-62d8-40e6-9966-6bdd666ce563"}`),
				),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "526b4319-62d8-40e6-9966-6bdd666ce563",
				Errors:   nil,
				Data: map[string]any{
					"id": "526b4319-62d8-40e6-9966-6bdd666ce563",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the Crm constituent emailaddresses",
			Input: common.WriteParams{ObjectName: "crm-conmg/emailaddresses", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm-conmg/emailaddresses"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(
					http.StatusOK, []byte(`{"id": "205334be-1c3e-45d7-be99-a17e5b48b159"}`),
				),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "205334be-1c3e-45d7-be99-a17e5b48b159",
				Errors:   nil,
				Data: map[string]any{
					"id": "205334be-1c3e-45d7-be99-a17e5b48b159",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Updating the Crm constituent emailaddresses",
			Input: common.WriteParams{
				ObjectName: "crm-conmg/emailaddresses",
				RecordData: "dummy",
				RecordId:   "205334be-1c3e-45d7-be99-a17e5b48b159",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm-conmg/emailaddresses/205334be-1c3e-45d7-be99-a17e5b48b159"),
					mockcond.MethodPATCH(),
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
