package sendgrid

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createListResp := testutils.DataFromFile(t, "write/list-create.json")
	updateListResp := testutils.DataFromFile(t, "write/list-update.json")
	createTemplateResp := testutils.DataFromFile(t, "write/template-create.json")
	createASMResp := testutils.DataFromFile(t, "write/asm-group-create.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectLists},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Unknown object is not supported",
			Input: common.WriteParams{
				ObjectName: objectBounces,
				RecordData: map[string]any{"email": "x@example.com"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create list",
			Input: common.WriteParams{
				ObjectName: objectLists,
				RecordData: map[string]any{
					"name": "amp-test-list",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/marketing/lists"),
					mockcond.Body(`{"name":"amp-test-list"}`),
				},
				Then: mockserver.Response(http.StatusCreated, createListResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ca7a3796-e8a8-4029-9ccb-df8937940562",
				Data: map[string]any{
					"id":   "ca7a3796-e8a8-4029-9ccb-df8937940562",
					"name": "amp-test-list",
				},
			},
		},
		{
			Name: "Update list",
			Input: common.WriteParams{
				ObjectName: objectLists,
				RecordId:   "ca7a3796-e8a8-4029-9ccb-df8937940562",
				RecordData: map[string]any{
					"name": "amp-test-list-updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/marketing/lists/ca7a3796-e8a8-4029-9ccb-df8937940562"),
					mockcond.Body(`{"name":"amp-test-list-updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateListResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ca7a3796-e8a8-4029-9ccb-df8937940562",
				Data: map[string]any{
					"id":   "ca7a3796-e8a8-4029-9ccb-df8937940562",
					"name": "amp-test-list-updated",
				},
			},
		},
		{
			Name: "Create template",
			Input: common.WriteParams{
				ObjectName: objectTemplates,
				RecordData: map[string]any{
					"name":       "amp-test-template",
					"generation": "dynamic",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/templates"),
					mockcond.Body(`{"generation":"dynamic","name":"amp-test-template"}`),
				},
				Then: mockserver.Response(http.StatusCreated, createTemplateResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "733ba07f-ead1-41fc-933a-3976baa23716",
				Data: map[string]any{
					"id":         "733ba07f-ead1-41fc-933a-3976baa23716",
					"name":       "amp-test-template",
					"generation": "dynamic",
				},
			},
		},
		{
			Name: "Create ASM group",
			Input: common.WriteParams{
				ObjectName: objectASMGroups,
				RecordData: map[string]any{
					"name":        "amp-test-asm-group",
					"description": "Created by connector unit test",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/asm/groups"),
					mockcond.Body(`{"description":"Created by connector unit test","name":"amp-test-asm-group"}`),
				},
				Then: mockserver.Response(http.StatusCreated, createASMResp),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "12345",
				Data: map[string]any{
					"id":          float64(12345),
					"name":        "amp-test-asm-group",
					"description": "Created by connector unit test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
