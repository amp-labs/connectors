package odoo

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

func TestWrite(t *testing.T) {
	t.Parallel()

	respCreate := testutils.DataFromFile(t, "write-crm-lead-create.json")
	respUpdate := testutils.DataFromFile(t, "write-crm-lead-update.json")

	tests := []testroutines.Write{
		{
			Name: "Create crm.lead successfully",
			Input: common.WriteParams{
				ObjectName: "crm.lead",
				RecordData: map[string]any{
					"name":         "New lead",
					"contact_name": "Contact",
					"email_from":   "lead@example.com",
					"phone":        "+1-555-0100",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/json/2/crm.lead/create"),
					mockcond.Body(`{"vals_list":[{"contact_name":"Contact","email_from":"lead@example.com","name":"New lead","phone":"+1-555-0100"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, respCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "142",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update crm.lead successfully",
			Input: common.WriteParams{
				ObjectName: "crm.lead",
				RecordId:   "104",
				RecordData: map[string]any{
					"name":  "Updated lead name",
					"phone": "+81-80-0000-0000",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/json/2/crm.lead/write"),
					mockcond.Body(`{"ids":[104],"vals":{"name":"Updated lead name","phone":"+81-80-0000-0000"}}`),
				},
				Then: mockserver.Response(http.StatusOK, respUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "104",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create passes record unchanged including nested keys",
			Input: common.WriteParams{
				ObjectName: "crm.lead",
				RecordData: map[string]any{
					"context": map[string]any{"lang": "fr_FR"},
					"name":    "Lead FR",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/json/2/crm.lead/create"),
					mockcond.Body(`{"vals_list":[{"context":{"lang":"fr_FR"},"name":"Lead FR"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, respCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "142",
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
