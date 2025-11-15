package recurly

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Unknown object requested",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"accounts", "subscriptions", "plans"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"accounts": {
						DisplayName: "Accounts",
						Fields: map[string]common.FieldMetadata{
							"address": {
								DisplayName:  "address",
								ValueType:    "other",
								ProviderType: "object",
							},
							"bill_date": {
								DisplayName:  "bill_date",
								ValueType:    "string",
								ProviderType: "string",
							},
							"bill_to": {
								DisplayName:  "bill_to",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "parent", DisplayValue: "parent"},
									{Value: "self", DisplayValue: "self"},
								},
							},
							"billing_info": {
								DisplayName:  "billing_info",
								ValueType:    "other",
								ProviderType: "object",
							},
							"cc_emails": {
								DisplayName:  "cc_emails",
								ValueType:    "string",
								ProviderType: "string",
							},
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"company": {
								DisplayName:  "company",
								ValueType:    "string",
								ProviderType: "string",
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"custom_fields": {
								DisplayName:  "custom_fields",
								ValueType:    "other",
								ProviderType: "array",
							},
							"deleted_at": {
								DisplayName:  "deleted_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"dunning_campaign_id": {
								DisplayName:  "dunning_campaign_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"entity_use_code": {
								DisplayName:  "entity_use_code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"exemption_certificate": {
								DisplayName:  "exemption_certificate",
								ValueType:    "string",
								ProviderType: "string",
							},
							"external_accounts": {
								DisplayName:  "external_accounts",
								ValueType:    "other",
								ProviderType: "array",
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"has_active_subscription": {
								DisplayName:  "has_active_subscription",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"has_canceled_subscription": {
								DisplayName:  "has_canceled_subscription",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"has_future_subscription": {
								DisplayName:  "has_future_subscription",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"has_live_subscription": {
								DisplayName:  "has_live_subscription",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"has_past_due_invoice": {
								DisplayName:  "has_past_due_invoice",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"has_paused_subscription": {
								DisplayName:  "has_paused_subscription",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"hosted_login_token": {
								DisplayName:  "hosted_login_token",
								ValueType:    "string",
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"invoice_template_id": {
								DisplayName:  "invoice_template_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"last_name": {
								DisplayName:  "last_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"object": {
								DisplayName:  "object",
								ValueType:    "string",
								ProviderType: "string",
							},
							"override_business_entity_id": {
								DisplayName:  "override_business_entity_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"parent_account_id": {
								DisplayName:  "parent_account_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"preferred_locale": {
								DisplayName:  "preferred_locale",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "da-DK", DisplayValue: "da-DK"},
									{Value: "de-CH", DisplayValue: "de-CH"},
									{Value: "de-DE", DisplayValue: "de-DE"},
									{Value: "en-AU", DisplayValue: "en-AU"},
									{Value: "en-CA", DisplayValue: "en-CA"},
									{Value: "en-GB", DisplayValue: "en-GB"},
									{Value: "en-IE", DisplayValue: "en-IE"},
									{Value: "en-NZ", DisplayValue: "en-NZ"},
									{Value: "en-US", DisplayValue: "en-US"},
									{Value: "es-ES", DisplayValue: "es-ES"},
									{Value: "es-MX", DisplayValue: "es-MX"},
									{Value: "es-US", DisplayValue: "es-US"},
									{Value: "fi-FI", DisplayValue: "fi-FI"},
									{Value: "fr-BE", DisplayValue: "fr-BE"},
									{Value: "fr-CA", DisplayValue: "fr-CA"},
									{Value: "fr-CH", DisplayValue: "fr-CH"},
									{Value: "fr-FR", DisplayValue: "fr-FR"},
									{Value: "hi-IN", DisplayValue: "hi-IN"},
									{Value: "it-IT", DisplayValue: "it-IT"},
									{Value: "ja-JP", DisplayValue: "ja-JP"},
									{Value: "ko-KR", DisplayValue: "ko-KR"},
									{Value: "nl-BE", DisplayValue: "nl-BE"},
									{Value: "nl-NL", DisplayValue: "nl-NL"},
									{Value: "pl-PL", DisplayValue: "pl-PL"},
									{Value: "pt-BR", DisplayValue: "pt-BR"},
									{Value: "pt-PT", DisplayValue: "pt-PT"},
									{Value: "ro-RO", DisplayValue: "ro-RO"},
									{Value: "ru-RU", DisplayValue: "ru-RU"},
									{Value: "sk-SK", DisplayValue: "sk-SK"},
									{Value: "sv-SE", DisplayValue: "sv-SE"},
									{Value: "tr-TR", DisplayValue: "tr-TR"},
									{Value: "zh-CN", DisplayValue: "zh-CN"},
								},
							},
							"preferred_time_zone": {
								DisplayName:  "preferred_time_zone",
								ValueType:    "string",
								ProviderType: "string",
							},
							"shipping_addresses": {
								DisplayName:  "shipping_addresses",
								ValueType:    "other",
								ProviderType: "array",
							},
							"state": {
								DisplayName:  "state",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "active", DisplayValue: "active"},
									{Value: "inactive", DisplayValue: "inactive"},
								},
							},
							"tax_exempt": {
								DisplayName:  "tax_exempt",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"updated_at": {
								DisplayName:  "updated_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"username": {
								DisplayName:  "username",
								ValueType:    "string",
								ProviderType: "string",
							},
							"vat_number": {
								DisplayName:  "vat_number",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"subscriptions": {
						DisplayName: "Subscriptions",
						Fields: map[string]common.FieldMetadata{
							"account": {
								DisplayName:  "account",
								ValueType:    "other",
								ProviderType: "object",
							},
							"action_result": {
								DisplayName:  "action_result",
								ValueType:    "other",
								ProviderType: "object",
							},
							"activated_at": {
								DisplayName:  "activated_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"active_invoice_id": {
								DisplayName:  "active_invoice_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"add_ons": {
								DisplayName:  "add_ons",
								ValueType:    "other",
								ProviderType: "array",
							},
							"add_ons_total": {
								DisplayName:  "add_ons_total",
								ValueType:    "other",
								ProviderType: "number",
							},
							"auto_renew": {
								DisplayName:  "auto_renew",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"bank_account_authorized_at": {
								DisplayName:  "bank_account_authorized_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"billing_info_id": {
								DisplayName:  "billing_info_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"business_entity_id": {
								DisplayName:  "business_entity_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"canceled_at": {
								DisplayName:  "canceled_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"collection_method": {
								DisplayName:  "collection_method",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "automatic", DisplayValue: "automatic"},
									{Value: "manual", DisplayValue: "manual"},
								},
							},
							"converted_at": {
								DisplayName:  "converted_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"coupon_redemptions": {
								DisplayName:  "coupon_redemptions",
								ValueType:    "other",
								ProviderType: "array",
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"currency": {
								DisplayName:  "currency",
								ValueType:    "string",
								ProviderType: "string",
							},
							"current_period_ends_at": {
								DisplayName:  "current_period_ends_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"current_period_started_at": {
								DisplayName:  "current_period_started_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"current_term_ends_at": {
								DisplayName:  "current_term_ends_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"current_term_started_at": {
								DisplayName:  "current_term_started_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"custom_fields": {
								DisplayName:  "custom_fields",
								ValueType:    "other",
								ProviderType: "array",
							},
							"customer_notes": {
								DisplayName:  "customer_notes",
								ValueType:    "string",
								ProviderType: "string",
							},
							"expiration_reason": {
								DisplayName:  "expiration_reason",
								ValueType:    "string",
								ProviderType: "string",
							},
							"expires_at": {
								DisplayName:  "expires_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"gateway_code": {
								DisplayName:  "gateway_code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"net_terms": {
								DisplayName:  "net_terms",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"net_terms_type": {
								DisplayName:  "net_terms_type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "net", DisplayValue: "net"},
									{Value: "eom", DisplayValue: "eom"},
								},
							},
							"object": {
								DisplayName:  "object",
								ValueType:    "string",
								ProviderType: "string",
							},
							"paused_at": {
								DisplayName:  "paused_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"pending_change": {
								DisplayName:  "pending_change",
								ValueType:    "other",
								ProviderType: "object",
							},
							"plan": {
								DisplayName:  "plan",
								ValueType:    "other",
								ProviderType: "object",
							},
							"po_number": {
								DisplayName:  "po_number",
								ValueType:    "string",
								ProviderType: "string",
							},
							"price_segment_id": {
								DisplayName:  "price_segment_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"quantity": {
								DisplayName:  "quantity",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"ramp_intervals": {
								DisplayName:  "ramp_intervals",
								ValueType:    "other",
								ProviderType: "array",
							},
							"remaining_billing_cycles": {
								DisplayName:  "remaining_billing_cycles",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"remaining_pause_cycles": {
								DisplayName:  "remaining_pause_cycles",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"renewal_billing_cycles": {
								DisplayName:  "renewal_billing_cycles",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"revenue_schedule_type": {
								DisplayName:  "revenue_schedule_type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "at_range_end", DisplayValue: "at_range_end"},
									{Value: "at_range_start", DisplayValue: "at_range_start"},
									{Value: "evenly", DisplayValue: "evenly"},
									{Value: "never", DisplayValue: "never"},
								},
							},
							"shipping": {
								DisplayName:  "shipping",
								ValueType:    "other",
								ProviderType: "object",
							},
							"started_with_gift": {
								DisplayName:  "started_with_gift",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"state": {
								DisplayName:  "state",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "active", DisplayValue: "active"},
									{Value: "canceled", DisplayValue: "canceled"},
									{Value: "expired", DisplayValue: "expired"},
									{Value: "failed", DisplayValue: "failed"},
									{Value: "future", DisplayValue: "future"},
									{Value: "paused", DisplayValue: "paused"},
								},
							},
							"subtotal": {
								DisplayName:  "subtotal",
								ValueType:    "other",
								ProviderType: "number",
							},
							"tax": {
								DisplayName:  "tax",
								ValueType:    "other",
								ProviderType: "number",
							},
							"tax_inclusive": {
								DisplayName:  "tax_inclusive",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"tax_info": {
								DisplayName:  "tax_info",
								ValueType:    "other",
								ProviderType: "object",
							},
							"terms_and_conditions": {
								DisplayName:  "terms_and_conditions",
								ValueType:    "string",
								ProviderType: "string",
							},
							"total": {
								DisplayName:  "total",
								ValueType:    "other",
								ProviderType: "number",
							},
							"total_billing_cycles": {
								DisplayName:  "total_billing_cycles",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"trial_ends_at": {
								DisplayName:  "trial_ends_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"trial_started_at": {
								DisplayName:  "trial_started_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"unit_amount": {
								DisplayName:  "unit_amount",
								ValueType:    "other",
								ProviderType: "number",
							},
							"updated_at": {
								DisplayName:  "updated_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"uuid": {
								DisplayName:  "uuid",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"plans": {
						DisplayName: "Plans",
						Fields: map[string]common.FieldMetadata{
							"accounting_code": {
								DisplayName:  "accounting_code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"allow_any_item_on_subscriptions": {
								DisplayName:  "allow_any_item_on_subscriptions",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"auto_renew": {
								DisplayName:  "auto_renew",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"avalara_service_type": {
								DisplayName:  "avalara_service_type",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"avalara_transaction_type": {
								DisplayName:  "avalara_transaction_type",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"currencies": {
								DisplayName:  "currencies",
								ValueType:    "other",
								ProviderType: "array",
							},
							"custom_fields": {
								DisplayName:  "custom_fields",
								ValueType:    "other",
								ProviderType: "array",
							},
							"deleted_at": {
								DisplayName:  "deleted_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"description": {
								DisplayName:  "description",
								ValueType:    "string",
								ProviderType: "string",
							},
							"dunning_campaign_id": {
								DisplayName:  "dunning_campaign_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"hosted_pages": {
								DisplayName:  "hosted_pages",
								ValueType:    "other",
								ProviderType: "object",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"interval_length": {
								DisplayName:  "interval_length",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"interval_unit": {
								DisplayName:  "interval_unit",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "days", DisplayValue: "days"},
									{Value: "months", DisplayValue: "months"},
								},
							},
							"liability_gl_account_id": {
								DisplayName:  "liability_gl_account_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"object": {
								DisplayName:  "object",
								ValueType:    "string",
								ProviderType: "string",
							},
							"performance_obligation_id": {
								DisplayName:  "performance_obligation_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"pricing_model": {
								DisplayName:  "pricing_model",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "fixed", DisplayValue: "fixed"},
									{Value: "ramp", DisplayValue: "ramp"},
								},
							},
							"ramp_intervals": {
								DisplayName:  "ramp_intervals",
								ValueType:    "other",
								ProviderType: "array",
							},
							"revenue_gl_account_id": {
								DisplayName:  "revenue_gl_account_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"revenue_schedule_type": {
								DisplayName:  "revenue_schedule_type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "at_range_end", DisplayValue: "at_range_end"},
									{Value: "at_range_start", DisplayValue: "at_range_start"},
									{Value: "evenly", DisplayValue: "evenly"},
									{Value: "never", DisplayValue: "never"},
								},
							},
							"setup_fee_accounting_code": {
								DisplayName:  "setup_fee_accounting_code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"setup_fee_liability_gl_account_id": {
								DisplayName:  "setup_fee_liability_gl_account_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"setup_fee_performance_obligation_id": {
								DisplayName:  "setup_fee_performance_obligation_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"setup_fee_revenue_gl_account_id": {
								DisplayName:  "setup_fee_revenue_gl_account_id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"setup_fee_revenue_schedule_type": {
								DisplayName:  "setup_fee_revenue_schedule_type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "at_range_end", DisplayValue: "at_range_end"},
									{Value: "at_range_start", DisplayValue: "at_range_start"},
									{Value: "evenly", DisplayValue: "evenly"},
									{Value: "never", DisplayValue: "never"},
								},
							},
							"setup_fees": {
								DisplayName:  "setup_fees",
								ValueType:    "other",
								ProviderType: "array",
							},
							"state": {
								DisplayName:  "state",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "active", DisplayValue: "active"},
									{Value: "inactive", DisplayValue: "inactive"},
								},
							},
							"tax_code": {
								DisplayName:  "tax_code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"tax_exempt": {
								DisplayName:  "tax_exempt",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"total_billing_cycles": {
								DisplayName:  "total_billing_cycles",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"trial_length": {
								DisplayName:  "trial_length",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"trial_requires_billing_info": {
								DisplayName:  "trial_requires_billing_info",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"trial_unit": {
								DisplayName:  "trial_unit",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "days", DisplayValue: "days"},
									{Value: "months", DisplayValue: "months"},
								},
							},
							"updated_at": {
								DisplayName:  "updated_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"vertex_transaction_type": {
								DisplayName:  "vertex_transaction_type",
								ValueType:    "string",
								ProviderType: "string",
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
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
