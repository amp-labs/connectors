package stripe

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
)

// Endpoints excluded from connector implementation.
//	* /apps/secrets/delete
//	* /tax/transactions/create_from_calculation
//	* /tax/transactions/create_reversal
//	* /test_helpers/issuing/transactions/create_force_capture
//	* /test_helpers/issuing/transactions/create_unlinked_refund

const (
	objectNameAccountLinks                = "account_links"
	objectNameAccountSessions             = "account_sessions"
	objectNameAccounts                    = "accounts"
	objectNameApplePayDomain              = "domains"
	objectNameAppSecrets                  = "secrets"
	objectNameBillingAlerts               = "alerts"
	objectNameCreditGrants                = "credit_grants"
	objectNameMeterEventAdjustments       = "meter_event_adjustments"
	objectNameMeterEvents                 = "meter_events"
	objectNameBillingMeters               = "meters"
	objectNameConfigurations              = "billing_portal/configurations"
	objectNameBillingPortalSessions       = "billing_portal/sessions"
	objectNameCharges                     = "charges"
	objectNameCheckoutSessions            = "checkout/sessions"
	objectNameOrders                      = "orders"
	objectNameCoupons                     = "coupons"
	objectNameCreditNotes                 = "credit_notes"
	objectNameCustomerSessions            = "customer_sessions"
	objectNameCustomers                   = "customers"
	objectNameDisputes                    = "disputes"
	objectNameFeatures                    = "features"
	objectNameEphemeralKeys               = "ephemeral_keys"
	objectNameFileLinks                   = "file_links"
	objectNameFiles                       = "files"
	objectNameFinancialSessions           = "financial_connections/sessions"
	objectNameRequests                    = "requests"
	objectNameVerificationSessions        = "verification_sessions"
	objectNameInvoiceItems                = "invoiceitems"
	objectNameInvoices                    = "invoices"
	objectNameInvoicesPreview             = "invoices/create_preview"
	objectNameAuthorizations              = "issuing/authorizations"
	objectNameCardholders                 = "cardholders"
	objectNameCards                       = "cards"
	objectNameIssuingDisputes             = "issuing/disputes"
	objectNamePersonalizationDesigns      = "personalization_designs"
	objectNameSettlements                 = "settlements"
	objectNameIssuingTokens               = "issuing/tokens"
	objectNameTransactions                = "issuing/transactions"
	objectNamePaymentIntents              = "payment_intents"
	objectNamePaymentLinks                = "payment_links"
	objectNamePaymentMethodConfigurations = "payment_method_configurations"
	objectNamePaymentMethodDomains        = "payment_method_domains"
	objectNamePaymentMethods              = "payment_methods"
	objectNamePayouts                     = "payouts"
	objectNamePlans                       = "plans"
	objectNamePrices                      = "prices"
	objectNameProducts                    = "products"
	objectNamePromotionCodes              = "promotion_codes"
	objectNameQuotes                      = "quotes"
	objectNameValueListItems              = "value_list_items"
	objectNameValueLists                  = "value_lists"
	objectNameRefunds                     = "refunds"
	objectNameReportRuns                  = "report_runs"
	objectNameSetupIntents                = "setup_intents"
	objectNameShippingRates               = "shipping_rates"
	objectNameSources                     = "sources"
	objectNameSubscriptionItems           = "subscription_items"
	objectNameSubscriptionSchedules       = "subscription_schedules"
	objectNameSubscriptions               = "subscriptions"
	objectNameTaxCalculations             = "calculations"
	objectNameRaxRegistrations            = "registrations"
	objectNameTaxSettings                 = "settings"
	objectNameTaxIDs                      = "tax_ids"
	objectNameTaxRates                    = "tax_rates"
	objectNameTerminalConfigurations      = "terminal/configurations"
	objectNameConnectionTokens            = "connection_tokens"
	objectNameLocations                   = "locations"
	objectNameReaders                     = "readers"
	objectNameTestConfirmationTokens      = "test_helpers/confirmation_tokens"
	objectNameTestAuthorizations          = "test_helpers/issuing/authorizations"
	objectNameTestSettlements             = "test_helpers/issuing/settlements"
	objectNameTestClocks                  = "test_clocks"
	objectNameTestOutboundPayments        = "test_helpers/treasury/outbound_payments"
	objectNameTestOutboundTransfers       = "test_helpers/treasury/outbound_transfers"
	objectNameTestReceivedCredits         = "test_helpers/treasury/received_credits"
	objectNameTestReceivedDebits          = "test_helpers/treasury/received_debits"
	objectNameTokens                      = "tokens"
	objectNameTopups                      = "topups"
	objectNameTransfers                   = "transfers"
	objectNameCreditReversals             = "credit_reversals"
	objectNameDebitReversals              = "debit_reversals"
	objectNameFinancialAccounts           = "financial_accounts"
	objectNameWebhookEndpoints            = "webhook_endpoints"

	// The READ endpoint exists, but according to the OpenAPI spec, at least one query parameter is required.
	// Since treasury APIs are not enabled, making a valid request for verification is not possible.
	// As a result, these objects are set to be write-only, which likely they are.
	objectNameInboundTransfers  = "inbound_transfers"
	objectNameOutboundPayments  = "outbound_payments"
	objectNameOutboundTransfers = "outbound_transfers"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByCreate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	"": datautils.NewSet(
		objectNameAccountLinks,
		objectNameAccountSessions,
		objectNameAccounts,
		objectNameApplePayDomain,
		objectNameAppSecrets,
		objectNameBillingAlerts,
		objectNameCreditGrants,
		objectNameMeterEventAdjustments,
		objectNameMeterEvents,
		objectNameBillingMeters,
		objectNameConfigurations,
		objectNameBillingPortalSessions,
		objectNameCharges,
		objectNameCheckoutSessions,
		objectNameOrders,
		objectNameCoupons,
		objectNameCreditNotes,
		objectNameCustomerSessions,
		objectNameCustomers,
		objectNameFeatures,
		objectNameEphemeralKeys,
		objectNameFileLinks,
		objectNameFiles,
		objectNameFinancialSessions,
		objectNameRequests,
		objectNameVerificationSessions,
		objectNameInvoiceItems,
		objectNameInvoices,
		objectNameInvoicesPreview,
		objectNameCardholders,
		objectNameCards,
		objectNameIssuingDisputes,
		objectNamePersonalizationDesigns,
		objectNamePaymentIntents,
		objectNamePaymentLinks,
		objectNamePaymentMethodConfigurations,
		objectNamePaymentMethodDomains,
		objectNamePaymentMethods,
		objectNamePayouts,
		objectNamePlans,
		objectNamePrices,
		objectNameProducts,
		objectNamePromotionCodes,
		objectNameQuotes,
		objectNameValueListItems,
		objectNameValueLists,
		objectNameRefunds,
		objectNameReportRuns,
		objectNameSetupIntents,
		objectNameShippingRates,
		objectNameSources,
		objectNameSubscriptionItems,
		objectNameSubscriptionSchedules,
		objectNameSubscriptions,
		objectNameTaxCalculations,
		objectNameTaxSettings,
		objectNameTaxIDs,
		objectNameTaxRates,
		objectNameTerminalConfigurations,
		objectNameConnectionTokens,
		objectNameLocations,
		objectNameReaders,
		objectNameTestConfirmationTokens,
		objectNameTestAuthorizations,
		objectNameTestSettlements,
		objectNameTestClocks,
		objectNameTestReceivedCredits,
		objectNameTestReceivedDebits,
		objectNameTokens,
		objectNameTopups,
		objectNameTransfers,
		objectNameCreditReversals,
		objectNameDebitReversals,
		objectNameFinancialAccounts,
		objectNameInboundTransfers,
		objectNameOutboundPayments,
		objectNameOutboundTransfers,
		objectNameWebhookEndpoints,
	),
}

var supportedObjectsByUpdate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	"": datautils.NewSet(
		objectNameAccounts,
		objectNameCreditGrants,
		objectNameBillingMeters,
		objectNameConfigurations,
		objectNameCharges,
		objectNameCheckoutSessions,
		objectNameOrders,
		objectNameCoupons,
		objectNameCreditNotes,
		objectNameCustomers,
		objectNameDisputes,
		objectNameFeatures,
		objectNameFileLinks,
		objectNameVerificationSessions,
		objectNameInvoiceItems,
		objectNameInvoices,
		objectNameAuthorizations,
		objectNameCardholders,
		objectNameCards,
		objectNameIssuingDisputes,
		objectNamePersonalizationDesigns,
		objectNameSettlements,
		objectNameIssuingTokens,
		objectNameTransactions,
		objectNamePaymentIntents,
		objectNamePaymentLinks,
		objectNamePaymentMethodConfigurations,
		objectNamePaymentMethodDomains,
		objectNamePaymentMethods,
		objectNamePayouts,
		objectNamePlans,
		objectNamePrices,
		objectNameProducts,
		objectNamePromotionCodes,
		objectNameQuotes,
		objectNameValueLists,
		objectNameRefunds,
		objectNameSetupIntents,
		objectNameShippingRates,
		objectNameSources,
		objectNameSubscriptionItems,
		objectNameSubscriptionSchedules,
		objectNameSubscriptions,
		objectNameRaxRegistrations,
		objectNameTaxRates,
		objectNameTerminalConfigurations,
		objectNameLocations,
		objectNameReaders,
		objectNameTestOutboundPayments,
		objectNameTestOutboundTransfers,
		objectNameTopups,
		objectNameTransfers,
		objectNameFinancialAccounts,
		objectNameWebhookEndpoints,
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	"": datautils.NewSet(
		objectNameAccounts,
		objectNameApplePayDomain,
		objectNameCoupons,
		objectNameCustomers,
		objectNameEphemeralKeys,
		objectNameInvoiceItems,
		objectNameInvoices,
		objectNamePlans,
		objectNameProducts,
		objectNameValueListItems,
		objectNameValueLists,
		objectNameSubscriptionItems,
		objectNameSubscriptions,
		objectNameTaxIDs,
		objectNameTerminalConfigurations,
		objectNameLocations,
		objectNameReaders,
		objectNameTestClocks,
		objectNameWebhookEndpoints,
	),
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	// Read and Write objects.
	objectNameApplePayDomain:         "/v1/apple_pay/domains",
	objectNameBillingAlerts:          "/v1/billing/alerts",
	objectNameBillingMeters:          "/v1/billing/meters",
	objectNameCardholders:            "/v1/issuing/cardholders",
	objectNameCards:                  "/v1/issuing/cards",
	objectNameCreditGrants:           "/v1/billing/credit_grants",
	objectNameFeatures:               "/v1/entitlements/features",
	objectNameFinancialAccounts:      "/v1/treasury/financial_accounts",
	objectNameLocations:              "/v1/terminal/locations",
	objectNameOrders:                 "/v1/climate/orders",
	objectNamePersonalizationDesigns: "/v1/issuing/personalization_designs",
	objectNameRaxRegistrations:       "/v1/tax/registrations",
	objectNameReaders:                "/v1/terminal/readers",
	objectNameReportRuns:             "/v1/reporting/report_runs",
	objectNameRequests:               "/v1/forwarding/requests",
	objectNameTestClocks:             "/v1/test_helpers/test_clocks",
	objectNameValueLists:             "/v1/radar/value_lists",
	objectNameVerificationSessions:   "/v1/identity/verification_sessions",
	// Write only
	objectNameAppSecrets:            "/v1/apps/secrets",
	objectNameConnectionTokens:      "/v1/terminal/connection_tokens",
	objectNameCreditReversals:       "/v1/treasury/credit_reversals",
	objectNameDebitReversals:        "/v1/treasury/debit_reversals",
	objectNameInboundTransfers:      "/v1/treasury/inbound_transfers",
	objectNameMeterEventAdjustments: "/v1/billing/meter_event_adjustments",
	objectNameMeterEvents:           "/v1/billing/meter_events",
	objectNameOutboundPayments:      "/v1/treasury/outbound_payments",
	objectNameOutboundTransfers:     "/v1/treasury/outbound_transfers",
	objectNameSettlements:           "/v1/issuing/settlements",
	objectNameTaxCalculations:       "/v1/tax/calculations",
	objectNameTaxSettings:           "/v1/tax/settings",
	objectNameValueListItems:        "/v1/radar/value_list_items",
},
	func(objectName string) (jsonPath string) {
		return "/v1/" + objectName
	},
)
