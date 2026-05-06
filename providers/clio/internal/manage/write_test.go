package manage

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWriteGroups(t *testing.T) {
	t.Parallel()

	responseCreate := testutils.DataFromFile(t, "write-groups-create.json")
	responseUpdate := testutils.DataFromFile(t, "write-groups-update.json")

	tests := []testroutines.Write{
		{
			Name: "Create group successfully",
			Input: common.WriteParams{
				ObjectName: "groups",
				RecordData: map[string]any{
					"name": "Scout Unit Group",
					"type": "AdhocGroup",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/api/v4/groups.json"),
					mockcond.Body(`{"data":{"name":"Scout Unit Group","type":"AdhocGroup"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "884001",
				Data: map[string]any{
					"id":   float64(884001),
					"name": "Scout Unit Group",
					"type": "AdhocGroup",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update group successfully",
			Input: common.WriteParams{
				ObjectName: "groups",
				RecordId:   "884001",
				RecordData: map[string]any{
					"name": "Scout Unit Group Updated",
					"type": "AdhocGroup",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPatch),
					mockcond.Path("/api/v4/groups/884001.json"),
					mockcond.Body(`{"data":{"name":"Scout Unit Group Updated","type":"AdhocGroup"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "884001",
				Data: map[string]any{
					"id":   float64(884001),
					"name": "Scout Unit Group Updated",
					"type": "AdhocGroup",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return newWriteTestAdapter(tt.Server.URL)
			})
		})
	}
}

func newWriteTestAdapter(serverURL string) (*Adapter, error) {
	adapter, err := NewAdapter(common.ConnectorParams{
		Module:              providers.ModuleClioManage,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "app.clio.com",
		Metadata: map[string]string{
			"region": "",
		},
	})
	if err != nil {
		return nil, err
	}

	adapter.SetUnitTestMockServerBaseURL(serverURL)

	return adapter, nil
}
