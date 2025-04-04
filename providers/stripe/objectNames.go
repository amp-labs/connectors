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
	objectNameConfigurations              = "configurations"
	objectNameBillingPortalSessions       = "billing_sessions"
	objectNameCharges                     = "charges"
	objectNameCheckoutSessions            = "sessions_checkout"
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
	objectNameFinancialSessions           = "financial_sessions"
	objectNameRequests                    = "requests"
	objectNameVerificationSessions        = "verification_sessions"
	objectNameInvoiceItems                = "invoiceitems"
	objectNameInvoices                    = "invoices"
	objectNameInvoicesPreview             = "invoices_preview"
	objectNameAuthorizations              = "authorizations"
	objectNameCardholders                 = "cardholders"
	objectNameCards                       = "cards"
	objectNameIssuingDisputes             = "disputes_issuing"
	objectNamePersonalizationDesigns      = "personalization_designs"
	objectNameSettlements                 = "settlements"
	objectNameIssuingTokens               = "issuing_tokens"
	objectNameTransactions                = "transactions_issuing"
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
	objectNameTerminalConfigurations      = "configurations_terminal"
	objectNameConnectionTokens            = "connection_tokens"
	objectNameLocations                   = "locations"
	objectNameReaders                     = "readers"
	objectNameTestConfirmationTokens      = "test_confirmation_tokens"
	objectNameTestAuthorizations          = "test_authorizations"
	objectNameTestSettlements             = "test_settlements"
	objectNameTestClocks                  = "test_clocks"
	objectNameTestOutboundPayments        = "test_outbound_payments"
	objectNameTestOutboundTransfers       = "test_outbound_transfers"
	objectNameTestReceivedCredits         = "test_received_credits"
	objectNameTestReceivedDebits          = "test_received_debits"
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
	common.ModuleRoot: datautils.NewSet(
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
	common.ModuleRoot: datautils.NewSet(
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
	common.ModuleRoot: datautils.NewSet(
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
	objectNameApplePayDomain:         "/apple_pay/domains",
	objectNameAppSecrets:             "/apps/secrets",
	objectNameBillingAlerts:          "/billing/alerts",
	objectNameCreditGrants:           "/billing/credit_grants",
	objectNameMeterEventAdjustments:  "/billing/meter_event_adjustments",
	objectNameMeterEvents:            "/billing/meter_events",
	objectNameBillingMeters:          "/billing/meters",
	objectNameConfigurations:         "/billing_portal/configurations",
	objectNameBillingPortalSessions:  "/billing_portal/sessions",
	objectNameCheckoutSessions:       "/checkout/sessions",
	objectNameOrders:                 "/climate/orders",
	objectNameFeatures:               "/entitlements/features",
	objectNameFinancialSessions:      "/financial_connections/sessions",
	objectNameRequests:               "/forwarding/requests",
	objectNameVerificationSessions:   "/identity/verification_sessions",
	objectNameInvoicesPreview:        "/invoices/create_preview",
	objectNameAuthorizations:         "/issuing/authorizations",
	objectNameCardholders:            "/issuing/cardholders",
	objectNameCards:                  "/issuing/cards",
	objectNameIssuingDisputes:        "/issuing/disputes",
	objectNamePersonalizationDesigns: "/issuing/personalization_designs",
	objectNameSettlements:            "/issuing/settlements",
	objectNameIssuingTokens:          "/issuing/tokens",
	objectNameTransactions:           "/issuing/transactions",
	objectNameValueListItems:         "/radar/value_list_items",
	objectNameValueLists:             "/radar/value_lists",
	objectNameReportRuns:             "/reporting/report_runs",
	objectNameTaxCalculations:        "/tax/calculations",
	objectNameRaxRegistrations:       "/tax/registrations",
	objectNameTaxSettings:            "/tax/settings",
	objectNameTerminalConfigurations: "/terminal/configurations",
	objectNameConnectionTokens:       "/terminal/connection_tokens",
	objectNameLocations:              "/terminal/locations",
	objectNameReaders:                "/terminal/readers",
	objectNameTestConfirmationTokens: "/test_helpers/confirmation_tokens",
	objectNameTestAuthorizations:     "/test_helpers/issuing/authorizations",
	objectNameTestSettlements:        "/test_helpers/issuing/settlements",
	objectNameTestClocks:             "/test_helpers/test_clocks",
	objectNameTestOutboundPayments:   "/test_helpers/treasury/outbound_payments",
	objectNameTestOutboundTransfers:  "/test_helpers/treasury/outbound_transfers",
	objectNameTestReceivedCredits:    "/test_helpers/treasury/received_credits",
	objectNameTestReceivedDebits:     "/test_helpers/treasury/received_debits",
	objectNameCreditReversals:        "/treasury/credit_reversals",
	objectNameDebitReversals:         "/treasury/debit_reversals",
	objectNameFinancialAccounts:      "/treasury/financial_accounts",
	objectNameInboundTransfers:       "/treasury/inbound_transfers",
	objectNameOutboundPayments:       "/treasury/outbound_payments",
	objectNameOutboundTransfers:      "/treasury/outbound_transfers",
},
	func(objectName string) (jsonPath string) {
		return objectName
	},
)
