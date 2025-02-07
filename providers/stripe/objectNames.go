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
	objectNameCheckoutSessions            = "checkout_sessions"
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
	objectNameIssuingDisputes             = "issuing_disputes"
	objectNamePersonalizationDesigns      = "personalization_designs"
	objectNameSettlements                 = "settlements"
	objectNameIssuingTokens               = "issuing_tokens"
	objectNameTransactions                = "transactions"
	objectNameLinkAccountSessions         = "link_account_sessions"
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
	objectNameTerminalConfigurations      = "terminal_configurations"
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
	objectNameInboundTransfers            = "inbound_transfers"
	objectNameOutboundPayments            = "outbound_payments"
	objectNameOutboundTransfers           = "outbound_transfers"
	objectNameWebhookEndpoints            = "webhook_endpoints"
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
		objectNameLinkAccountSessions,
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
	objectNameApplePayDomain:         "/v1/apple_pay/domains",
	objectNameAppSecrets:             "/v1/apps/secrets",
	objectNameBillingAlerts:          "/v1/billing/alerts",
	objectNameCreditGrants:           "/v1/billing/credit_grants",
	objectNameMeterEventAdjustments:  "/v1/billing/meter_event_adjustments",
	objectNameMeterEvents:            "/v1/billing/meter_events",
	objectNameBillingMeters:          "/v1/billing/meters",
	objectNameConfigurations:         "/v1/billing_portal/configurations",
	objectNameBillingPortalSessions:  "/v1/billing_portal/sessions",
	objectNameCheckoutSessions:       "/v1/checkout/sessions",
	objectNameOrders:                 "/v1/climate/orders",
	objectNameFeatures:               "/v1/entitlements/features",
	objectNameFinancialSessions:      "/v1/financial_connections/sessions",
	objectNameRequests:               "/v1/forwarding/requests",
	objectNameVerificationSessions:   "/v1/identity/verification_sessions",
	objectNameInvoicesPreview:        "/v1/invoices/create_preview",
	objectNameAuthorizations:         "/v1/issuing/authorizations/",
	objectNameCardholders:            "/v1/issuing/cardholders",
	objectNameCards:                  "/v1/issuing/cards",
	objectNameIssuingDisputes:        "/v1/issuing/disputes",
	objectNamePersonalizationDesigns: "/v1/issuing/personalization_designs",
	objectNameSettlements:            "/v1/issuing/settlements",
	objectNameIssuingTokens:          "/v1/issuing/tokens",
	objectNameTransactions:           "/v1/issuing/transactions",
	objectNameValueListItems:         "/v1/radar/value_list_items",
	objectNameValueLists:             "/v1/radar/value_lists",
	objectNameReportRuns:             "/v1/reporting/report_runs",
	objectNameTaxCalculations:        "/v1/tax/calculations",
	objectNameRaxRegistrations:       "/v1/tax/registrations",
	objectNameTaxSettings:            "/v1/tax/settings",
	objectNameTerminalConfigurations: "/v1/terminal/configurations",
	objectNameConnectionTokens:       "/v1/terminal/connection_tokens",
	objectNameLocations:              "/v1/terminal/locations",
	objectNameReaders:                "/v1/terminal/readers",
	objectNameTestConfirmationTokens: "/v1/test_helpers/confirmation_tokens",
	objectNameTestAuthorizations:     "/v1/test_helpers/issuing/authorizations",
	objectNameTestSettlements:        "/v1/test_helpers/issuing/settlements",
	objectNameTestClocks:             "/v1/test_helpers/test_clocks",
	objectNameTestOutboundPayments:   "/v1/test_helpers/treasury/outbound_payments",
	objectNameTestOutboundTransfers:  "/v1/test_helpers/treasury/outbound_transfers",
	objectNameTestReceivedCredits:    "/v1/test_helpers/treasury/received_credits",
	objectNameTestReceivedDebits:     "/v1/test_helpers/treasury/received_debits",
	objectNameCreditReversals:        "/v1/treasury/credit_reversals",
	objectNameDebitReversals:         "/v1/treasury/debit_reversals",
	objectNameFinancialAccounts:      "/v1/treasury/financial_accounts",
	objectNameInboundTransfers:       "/v1/treasury/inbound_transfers",
	objectNameOutboundPayments:       "/v1/treasury/outbound_payments",
	objectNameOutboundTransfers:      "/v1/treasury/outbound_transfers",
},
	func(objectName string) (jsonPath string) {
		return "/v1/" + objectName
	},
)
