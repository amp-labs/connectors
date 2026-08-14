package webhook

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// eventDefinition contains the Stripe webhook event names associated with
// generic create, update, and delete operations for a resource.
//
// An empty field indicates that Stripe does not provide a webhook event suitable for that operation.
type eventDefinition struct {
	create string
	update string
	delete string
}

// eventDefinitions maps connector object names to their supported Stripe
// webhook event definitions.
// Doc URL: https://docs.stripe.com/api/events/types
var eventDefinitions = map[string]eventDefinition{ // nolint:gochecknoglobals
	"accounts": {
		update: "account.updated",
	},
	"application_fees": {
		create: "application_fee.created",
		// "application_fee.refunded",
	},
	// Object doesn't have CUD mappable event types.
	// "billing/alerts": {
	//	"billing.alert.triggered",
	// },
	"billing/credit_balance_transactions": {
		create: "billing.credit_balance_transaction.created",
	},
	"billing/credit_grants": {
		create: "billing.credit_grant.created",
		update: "billing.credit_grant.updated",
	},
	"billing/meters": {
		create: "billing.meter.created",
		update: "billing.meter.updated",
		// "billing.meter.deactivated",
		// "billing.meter.reactivated",
	},
	"billing_portal/configurations": {
		create: "billing_portal.configuration.created",
		update: "billing_portal.configuration.updated",
	},
	"charges": {
		update: "charge.updated",
		// "charge.captured",
		// "charge.expired",
		// "charge.failed",
		// "charge.pending",
		// "charge.refunded",
		// "charge.succeeded",
	},
	"checkout/sessions": {
		delete: "checkout.session.expired",
		// "checkout.session.async_payment_failed",
		// "checkout.session.async_payment_succeeded",
		// "checkout.session.completed",
	},
	"climate/orders": {
		create: "climate.order.created",
		delete: "climate.order.canceled",
		// "climate.order.delayed",
		// "climate.order.delivered",
		// "climate.order.product_substituted",
	},
	"climate/products": {
		create: "climate.product.created",
		update: "climate.product.pricing_updated",
	},
	"coupons": {
		create: "coupon.created",
		delete: "coupon.deleted",
		update: "coupon.updated",
	},
	"credit_notes": {
		create: "credit_note.created",
		delete: "credit_note.voided",
		update: "credit_note.updated",
	},
	"customers": {
		create: "customer.created",
		delete: "customer.deleted",
		update: "customer.updated",
	},
	"disputes": {
		create: "charge.dispute.created",
		delete: "charge.dispute.closed",
		update: "charge.dispute.updated",
		// "charge.dispute.funds_reinstated",
		// "charge.dispute.funds_withdrawn",
	},
	"files": {
		create: "file.created",
	},
	"financial_connections/accounts": {
		create: "financial_connections.account.created",
		// "financial_connections.account.account_numbers_updated",
		// "financial_connections.account.deactivated",
		// "financial_connections.account.disconnected",
		// "financial_connections.account.expected_deactivation_date_updated",
		// "financial_connections.account.reactivated",
		// "financial_connections.account.refreshed_balance",
		// "financial_connections.account.refreshed_ownership",
		// "financial_connections.account.refreshed_transactions",
		// "financial_connections.account.supported_payment_method_types_updated",
		// "financial_connections.account.upcoming_account_number_expiry",
		// "financial_connections.account.upcoming_deactivation",
	},
	"identity/verification_sessions": {
		delete: "identity.verification_session.canceled",
		create: "identity.verification_session.created",
		// "identity.verification_session.processing",
		// "identity.verification_session.redacted",
		// "identity.verification_session.requires_input",
		// "identity.verification_session.verified",
	},
	"invoice_payments": {
		create: "invoice_payment.paid",
	},
	"invoiceitems": {
		create: "invoiceitem.created",
		delete: "invoiceitem.deleted",
	},
	"invoices": {
		create: "invoice.created",
		delete: "invoice.deleted",
		update: "invoice.updated",
		// "invoice.finalization_failed",
		// "invoice.finalized",
		// "invoice.marked_uncollectible",
		// "invoice.overdue",
		// "invoice.overpaid",
		// "invoice.paid",
		// "invoice.payment_action_required",
		// "invoice.payment_attempt_required",
		// "invoice.payment_failed",
		// "invoice.payment_succeeded",
		// "invoice.sent",
		// "invoice.upcoming",
		// "invoice.voided",
		// "invoice.will_be_due",
	},
	"issuing/authorizations": {
		create: "issuing_authorization.created",
		update: "issuing_authorization.updated",
		// "issuing_authorization.request",
	},
	"issuing/cardholders": {
		create: "issuing_cardholder.created",
		update: "issuing_cardholder.updated",
	},
	"issuing/cards": {
		create: "issuing_card.created",
		update: "issuing_card.updated",
	},
	"issuing/disputes": {
		create: "issuing_dispute.created",
		update: "issuing_dispute.updated",
		// "issuing_dispute.closed",
		// "issuing_dispute.funds_reinstated",
		// "issuing_dispute.funds_rescinded",
		// "issuing_dispute.submitted",
	},
	"issuing/personalization_designs": {
		update: "issuing_personalization_design.updated",
		// "issuing_personalization_design.activated",
		// "issuing_personalization_design.deactivated",
		// "issuing_personalization_design.rejected",
	},
	"issuing/transactions": {
		create: "issuing_transaction.created",
		update: "issuing_transaction.updated",
		// "issuing_transaction.purchase_details_receipt_updated",
	},
	"payment_intents": {
		create: "payment_intent.created",
		delete: "payment_intent.canceled",
		// "payment_intent.amount_capturable_updated",
		// "payment_intent.partially_funded",
		// "payment_intent.payment_failed",
		// "payment_intent.processing",
		// "payment_intent.requires_action",
		// "payment_intent.succeeded",
	},
	"payment_links": {
		create: "payment_link.created",
		update: "payment_link.updated",
	},
	"payment_methods": {
		update: "payment_method.updated",
		// "payment_method.attached",
		// "payment_method.automatically_updated",
		// "payment_method.detached",
	},
	"payouts": {
		create: "payout.created",
		delete: "payout.canceled",
		update: "payout.updated",
		// "payout.failed",
		// "payout.paid",
		// "payout.reconciliation_completed",
	},
	"plans": {
		create: "plan.created",
		delete: "plan.deleted",
		update: "plan.updated",
	},
	"prices": {
		create: "price.created",
		delete: "price.deleted",
		update: "price.updated",
	},
	"products": {
		create: "product.created",
		delete: "product.deleted",
		update: "product.updated",
	},
	"promotion_codes": {
		create: "promotion_code.created",
		update: "promotion_code.updated",
	},
	"quotes": {
		create: "quote.created",
		delete: "quote.canceled",
		// "quote.accepted",
		// "quote.finalized",
		// "quote.will_expire",
	},
	"radar/early_fraud_warnings": {
		create: "radar.early_fraud_warning.created",
		update: "radar.early_fraud_warning.updated",
	},
	"refunds": {
		create: "refund.created",
		update: "refund.updated",
		// "refund.failed",
	},
	// Object doesn't have CUD mappable event types.
	// "reporting/report_runs": {
	//	"reporting.report_run.failed",
	// 	"reporting.report_run.succeeded",
	// },
	"reporting/report_types": {
		update: "reporting.report_type.updated",
	},
	"reviews": {
		delete: "review.closed",
		create: "review.opened",
	},
	"setup_intents": {
		create: "setup_intent.created",
		delete: "setup_intent.canceled",
		// "setup_intent.requires_action",
		// "setup_intent.setup_failed",
		// "setup_intent.succeeded",
	},
	"sigma/scheduled_query_runs": {
		create: "sigma.scheduled_query_run.created",
	},
	"subscription_schedules": {
		create: "subscription_schedule.created",
		delete: "subscription_schedule.canceled",
		update: "subscription_schedule.updated",
		// "subscription_schedule.aborted",
		// "subscription_schedule.completed",
		// "subscription_schedule.expiring",
		// "subscription_schedule.released",
	},
	"subscriptions": {
		create: "customer.subscription.created",
		delete: "customer.subscription.deleted",
		update: "customer.subscription.updated",
		// "customer.subscription.paused",
		// "customer.subscription.pending_update_applied",
		// "customer.subscription.pending_update_expired",
		// "customer.subscription.resumed",
		// "customer.subscription.trial_will_end",
	},
	"tax_ids": {
		create: "customer.tax_id.created",
		delete: "customer.tax_id.deleted",
		update: "customer.tax_id.updated",
	},
	"tax_rates": {
		create: "tax_rate.created",
		update: "tax_rate.updated",
	},
	"terminal/readers": {
		update: "terminal.reader.action_updated",
		// "terminal.reader.action_failed",
		// "terminal.reader.action_succeeded",
	},
	"test_helpers/test_clocks": {
		create: "test_helpers.test_clock.created",
		delete: "test_helpers.test_clock.deleted",
		// "test_helpers.test_clock.advancing",
		// "test_helpers.test_clock.internal_failure",
		// "test_helpers.test_clock.ready",
	},
	"topups": {
		create: "topup.created",
		delete: "topup.canceled",
		// "topup.failed",
		// "topup.reversed",
		// "topup.succeeded",
	},
	"transfers": {
		create: "transfer.created",
		update: "transfer.updated",
		// "transfer.reversed"
	},
}

// GetEventName returns the Stripe webhook event type associated with eventType
// for objectName.
//
// It returns common.ErrObjectNotSupported when objectName is not registered.
// It returns common.ErrSubscribeEventNotSupportedForObject when the object is
// registered but does not support the requested subscription operation.
func GetEventName(eventType common.SubscriptionEventType, objectName common.ObjectName) (string, error) {
	definition, ok := eventDefinitions[objectName.String()]
	if !ok {
		return "", fmt.Errorf("%w: object %v", common.ErrObjectNotSupported, objectName)
	}

	switch eventType { // nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		if definition.create != "" {
			return definition.create, nil
		}
	case common.SubscriptionEventTypeUpdate:
		if definition.update != "" {
			return definition.update, nil
		}
	case common.SubscriptionEventTypeDelete:
		if definition.delete != "" {
			return definition.delete, nil
		}
	}

	return "", fmt.Errorf("%w: event %s", common.ErrSubscribeEventNotSupportedForObject, eventType)
}
