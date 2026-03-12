package revenuecat

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

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	respApp := testutils.DataFromFile(t, "write/app.json")
	respCustomer := testutils.DataFromFile(t, "write/customer.json")
	respEntitlement := testutils.DataFromFile(t, "write/entitlement.json")
	respWebhook := testutils.DataFromFile(t, "write/integration-webhook.json")
	respOffering := testutils.DataFromFile(t, "write/offering.json")
	respProduct := testutils.DataFromFile(t, "write/product.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write data must be included",
			Input:        common.WriteParams{ObjectName: "entitlements"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unsupported object is rejected",
			Input:        common.WriteParams{ObjectName: "subscriptions", RecordData: map[string]any{"x": "y"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create app",
			Input: common.WriteParams{
				ObjectName: "apps",
				RecordData: map[string]any{"name": "My iOS App", "type": "app_store"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/projects/proj_123/apps"),
				},
				Then: mockserver.Response(http.StatusCreated, respApp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "app_xyz123",
				Data:     map[string]any{"id": "app_xyz123", "name": "My iOS App"},
			},
		},
		{
			Name: "Update app",
			Input: common.WriteParams{
				ObjectName: "apps",
				RecordId:   "app_xyz123",
				RecordData: map[string]any{"name": "My iOS App"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/projects/proj_123/apps/app_xyz123"),
				},
				Then: mockserver.Response(http.StatusOK, respApp),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "app_xyz123",
				Data:     map[string]any{"id": "app_xyz123", "name": "My iOS App"},
			},
		},
		{
			// Customers are created by the mobile SDK; only updates are supported via the REST API.
			Name: "Update customer",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "cust_abc456",
				RecordData: map[string]any{"attributes": map[string]any{}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/projects/proj_123/customers/cust_abc456"),
				},
				Then: mockserver.Response(http.StatusOK, respCustomer),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cust_abc456",
				Data:     map[string]any{"id": "cust_abc456"},
			},
		},
		{
			Name: "Create entitlement",
			Input: common.WriteParams{
				ObjectName: "entitlements",
				RecordData: map[string]any{"lookup_key": "premium", "display_name": "Premium Access"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/projects/proj_123/entitlements"),
				},
				Then: mockserver.Response(http.StatusCreated, respEntitlement),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "entl_abc123",
				Data:     map[string]any{"id": "entl_abc123", "lookup_key": "premium"},
			},
		},
		{
			Name: "Update entitlement",
			Input: common.WriteParams{
				ObjectName: "entitlements",
				RecordId:   "entl_abc123",
				RecordData: map[string]any{"display_name": "Premium Access"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/projects/proj_123/entitlements/entl_abc123"),
				},
				Then: mockserver.Response(http.StatusOK, respEntitlement),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "entl_abc123",
				Data:     map[string]any{"id": "entl_abc123", "lookup_key": "premium"},
			},
		},
		{
			Name: "Create integration webhook",
			Input: common.WriteParams{
				ObjectName: "integrations_webhooks",
				RecordData: map[string]any{"name": "My Webhook", "url": "https://example.com/webhook"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/projects/proj_123/integrations_webhooks"),
				},
				Then: mockserver.Response(http.StatusCreated, respWebhook),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "whi_xyz789",
				Data:     map[string]any{"id": "whi_xyz789", "name": "My Webhook"},
			},
		},
		{
			Name: "Update integration webhook",
			Input: common.WriteParams{
				ObjectName: "integrations_webhooks",
				RecordId:   "whi_xyz789",
				RecordData: map[string]any{"name": "My Webhook"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/projects/proj_123/integrations_webhooks/whi_xyz789"),
				},
				Then: mockserver.Response(http.StatusOK, respWebhook),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "whi_xyz789",
				Data:     map[string]any{"id": "whi_xyz789", "name": "My Webhook"},
			},
		},
		{
			Name: "Create offering",
			Input: common.WriteParams{
				ObjectName: "offerings",
				RecordData: map[string]any{"lookup_key": "default", "display_name": "Default Offering"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/projects/proj_123/offerings"),
				},
				Then: mockserver.Response(http.StatusCreated, respOffering),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ofng_abc789",
				Data:     map[string]any{"id": "ofng_abc789", "lookup_key": "default"},
			},
		},
		{
			Name: "Update offering",
			Input: common.WriteParams{
				ObjectName: "offerings",
				RecordId:   "ofng_abc789",
				RecordData: map[string]any{"display_name": "Default Offering"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/projects/proj_123/offerings/ofng_abc789"),
				},
				Then: mockserver.Response(http.StatusOK, respOffering),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ofng_abc789",
				Data:     map[string]any{"id": "ofng_abc789", "lookup_key": "default"},
			},
		},
		{
			// Products do not support PATCH; only create and delete are available.
			Name: "Create product",
			Input: common.WriteParams{
				ObjectName: "products",
				RecordData: map[string]any{"store_identifier": "monthly.premium", "type": "subscription"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/projects/proj_123/products"),
				},
				Then: mockserver.Response(http.StatusCreated, respProduct),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "prod_xyz789",
				Data:     map[string]any{"id": "prod_xyz789", "store_identifier": "monthly.premium"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestReadConnector(tt.Server.URL, "proj_123")
			})
		})
	}
}
