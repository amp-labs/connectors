package reader

import "github.com/amp-labs/connectors/internal/datautils"

// incrementalObjects contains object names that support incremental reading
// via the "created[gte]" query parameter.
var incrementalObjects = datautils.NewSet( // nolint:gochecknoglobals
	"accounts",
	"application_fees",
	"balance/history",
	"balance_transactions",
	"charges",
	"checkout/sessions",
	"coupons",
	"credit_notes",
	"customers",
	"disputes",
	"events",
	"file_links",
	"files",
	"forwarding/requests",
	"identity/verification_reports",
	"identity/verification_sessions",
	"invoiceitems",
	"invoices",
	"issuing/authorizations",
	"issuing/cardholders",
	"issuing/cards",
	"issuing/disputes",
	"issuing/transactions",
	"payment_intents",
	"payouts",
	"plans",
	"prices",
	"products",
	"promotion_codes",
	"radar/early_fraud_warnings",
	"radar/value_lists",
	"refunds",
	"reporting/report_runs",
	"reviews",
	"setup_intents",
	"shipping_rates",
	"subscription_schedules",
	"subscriptions",
	"tax_rates",
	"topups",
	"transfers",
	"treasury/financial_accounts",
)

// scopedObjectsForFinancialAccount lists Treasury objects that are context-scoped by financial_account.
//
// Requests to these endpoints must be resolved against a FinancialAccount,
// which is discovered through treasury/financial_accounts.
// The connector uses those accounts as context to list the corresponding objects.
var scopedObjectsForFinancialAccount = datautils.NewSet( // nolint:gochecknoglobals
	"treasury/outbound_payments",
	"treasury/received_credits",
	"treasury/debit_reversals",
	"treasury/inbound_transfers",
	"treasury/received_debits",
	"treasury/transaction_entries",
	"treasury/outbound_transfers",
	"treasury/transactions",
	"treasury/credit_reversals",
)
