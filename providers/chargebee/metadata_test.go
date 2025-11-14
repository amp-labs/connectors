package chargebee

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCustomers := testutils.DataFromFile(t, "customers.json")
	responseSubscriptions := testutils.DataFromFile(t, "subscriptions.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe one object with metadata",
			Input: []string{"customers"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/customers"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomers),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"allow_direct_debit": {
								DisplayName: "allow_direct_debit",
								ValueType:   common.ValueTypeBoolean,
							},
							"auto_collection": {
								DisplayName: "auto_collection",
								ValueType:   common.ValueTypeString,
							},
							"card_status": {
								DisplayName: "card_status",
								ValueType:   common.ValueTypeString,
							},
							"created_at": {
								DisplayName: "created_at",
								ValueType:   common.ValueTypeFloat,
							},
							"deleted": {
								DisplayName: "deleted",
								ValueType:   common.ValueTypeBoolean,
							},
							"email": {
								DisplayName: "email",
								ValueType:   common.ValueTypeString,
							},
							"excess_payments": {
								DisplayName: "excess_payments",
								ValueType:   common.ValueTypeFloat,
							},
							"first_name": {
								DisplayName: "first_name",
								ValueType:   common.ValueTypeString,
							},
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"last_name": {
								DisplayName: "last_name",
								ValueType:   common.ValueTypeString,
							},
							"net_term_days": {
								DisplayName: "net_term_days",
								ValueType:   common.ValueTypeFloat,
							},
							"object": {
								DisplayName: "object",
								ValueType:   common.ValueTypeString,
							},
							"pii_cleared": {
								DisplayName: "pii_cleared",
								ValueType:   common.ValueTypeString,
							},
							"preferred_currency_code": {
								DisplayName: "preferred_currency_code",
								ValueType:   common.ValueTypeString,
							},
							"promotional_credits": {
								DisplayName: "promotional_credits",
								ValueType:   common.ValueTypeFloat,
							},
							"refundable_credits": {
								DisplayName: "refundable_credits",
								ValueType:   common.ValueTypeFloat,
							},
							"resource_version": {
								DisplayName: "resource_version",
								ValueType:   common.ValueTypeFloat,
							},
							"taxability": {
								DisplayName: "taxability",
								ValueType:   common.ValueTypeString,
							},
							"unbilled_charges": {
								DisplayName: "unbilled_charges",
								ValueType:   common.ValueTypeFloat,
							},
							"updated_at": {
								DisplayName: "updated_at",
								ValueType:   common.ValueTypeFloat,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe subscriptions object",
			Input: []string{"subscriptions"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/subscriptions"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseSubscriptions),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"subscriptions": {
						DisplayName: "Subscriptions",
						Fields: map[string]common.FieldMetadata{
							"activated_at": {
								DisplayName: "activated_at",
								ValueType:   common.ValueTypeFloat,
							},
							"billing_period": {
								DisplayName: "billing_period",
								ValueType:   common.ValueTypeFloat,
							},
							"billing_period_unit": {
								DisplayName: "billing_period_unit",
								ValueType:   common.ValueTypeString,
							},
							"created_at": {
								DisplayName: "created_at",
								ValueType:   common.ValueTypeFloat,
							},
							"currency_code": {
								DisplayName: "currency_code",
								ValueType:   common.ValueTypeString,
							},
							"current_term_end": {
								DisplayName: "current_term_end",
								ValueType:   common.ValueTypeFloat,
							},
							"current_term_start": {
								DisplayName: "current_term_start",
								ValueType:   common.ValueTypeFloat,
							},
							"customer_id": {
								DisplayName: "customer_id",
								ValueType:   common.ValueTypeString,
							},
							"deleted": {
								DisplayName: "deleted",
								ValueType:   common.ValueTypeBoolean,
							},
							"due_invoices_count": {
								DisplayName: "due_invoices_count",
								ValueType:   common.ValueTypeFloat,
							},
							"due_since": {
								DisplayName: "due_since",
								ValueType:   common.ValueTypeFloat,
							},
							"has_scheduled_changes": {
								DisplayName: "has_scheduled_changes",
								ValueType:   common.ValueTypeBoolean,
							},
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"mrr": {
								DisplayName: "mrr",
								ValueType:   common.ValueTypeFloat,
							},
							"next_billing_at": {
								DisplayName: "next_billing_at",
								ValueType:   common.ValueTypeFloat,
							},
							"object": {
								DisplayName: "object",
								ValueType:   common.ValueTypeString,
							},
							"remaining_billing_cycles": {
								DisplayName: "remaining_billing_cycles",
								ValueType:   common.ValueTypeFloat,
							},
							"resource_version": {
								DisplayName: "resource_version",
								ValueType:   common.ValueTypeFloat,
							},
							"started_at": {
								DisplayName: "started_at",
								ValueType:   common.ValueTypeFloat,
							},
							"status": {
								DisplayName: "status",
								ValueType:   common.ValueTypeString,
							},
							"subscription_items": {
								DisplayName: "subscription_items",
								ValueType:   common.ValueTypeOther,
							},
							"total_dues": {
								DisplayName: "total_dues",
								ValueType:   common.ValueTypeFloat,
							},
							"updated_at": {
								DisplayName: "updated_at",
								ValueType:   common.ValueTypeFloat,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe multiple objects",
			Input: []string{"customers", "subscriptions"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/api/v2/customers"),
							mockcond.QueryParam("limit", "1"),
						},
						Then: mockserver.Response(http.StatusOK, responseCustomers),
					},
					{
						If: mockcond.And{
							mockcond.Path("/api/v2/subscriptions"),
							mockcond.QueryParam("limit", "1"),
						},
						Then: mockserver.Response(http.StatusOK, responseSubscriptions),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
						},
					},
					"subscriptions": {
						DisplayName: "Subscriptions",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "withampersand-demo-test",
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
