package reply

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Reply delivers one JSON document per event to the subscription URL. The event type is
// only in the body, as Reply sends no event-name or signature header (verified live).
// Every event is an activity (an email was sent, a contact replied, ...), never a create,
// update or delete of a record, so they are all pass-through events, each tied to the
// record it concerns: a contact, an email account, a LinkedIn account or a sequence.
//
// https://docs.reply.io/webhook-events
// https://docs.reply.io/webhook-event-payloads

// SubscriptionEvent is the parsed Reply webhook payload.
//
// Sample delivery (contact_opted_out):
//
//	{
//	  "event": {
//	    "date": "2026-10-30T14:05:55.4146886Z",
//	    "id": "bc1c2b71-bfd4-42b1-8955-e064895c3cb3",
//	    "type": "contact_opted_out",
//	    "user_id": 511180,
//	    "team_id": 100,
//	    "subscription_id": 24880
//	  },
//	  "opt_out_date": "2026-10-30T14:05:55.410816Z",
//	  "contact_fields": {"id": 12345, "email": "test-prospect@example.com"},
//	  "sequence_fields": {"id": 200, "name": "Test Sequence", "step_number": 1}
//	}
type SubscriptionEvent map[string]any

// CollapsedSubscriptionEvent is the raw Reply webhook payload as received. Reply delivers
// exactly one event per request, so the collapsed form expands to a one-element list.
type CollapsedSubscriptionEvent map[string]any

var (
	_ common.SubscriptionEvent            = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent   = CollapsedSubscriptionEvent{}
	_ connectors.WebhookVerifierConnector = &Connector{}
)

const (
	objectContacts        = "contacts"
	objectEmailAccounts   = "email-accounts"
	objectLinkedInAccount = "linkedin-accounts"
	objectSequences       = "sequences"

	eventFieldEnvelope = "event"
	eventFieldType     = "type"
	eventFieldDate     = "date"
)

var (
	errMissingEventType    = errors.New("reply event: missing event.type")
	errUnsupportedEvent    = errors.New("reply event: event type is not supported")
	errMissingRecordID     = errors.New("reply event: missing record id in payload")
	errMissingEventDate    = errors.New("reply event: missing event.date")
	errUnparsableEventDate = errors.New("reply event: unparsable event.date")
)

// eventRoute says which object an event concerns and where its record id sits in the payload.
type eventRoute struct {
	object string
	// idPath is the payload path to the record id.
	idPath []string
}

// The routes events share, by the object whose record they concern.
//
//nolint:gochecknoglobals
var (
	contactRoute         = eventRoute{object: objectContacts, idPath: []string{"contact_fields", "id"}}
	emailAccountRoute    = eventRoute{object: objectEmailAccounts, idPath: []string{"email_account_id"}}
	linkedInAccountRoute = eventRoute{object: objectLinkedInAccount, idPath: []string{"linkedin_id"}}
	sequenceRoute        = eventRoute{object: objectSequences, idPath: []string{"sequence_fields", "id"}}
)

// eventRoutes lists every event type Reply can send, as returned by GET /v3/webhooks/events.
//
//nolint:gochecknoglobals
var eventRoutes = map[string]eventRoute{
	// Email engagement.
	"email_sent":         contactRoute,
	"email_opened":       contactRoute,
	"email_link_clicked": contactRoute,
	"email_replied":      contactRoute,
	"reply_categorized":  contactRoute,
	"email_bounced":      contactRoute,
	"email_auto_reply":   contactRoute,
	// Email account health.
	"email_account_connection_lost": emailAccountRoute,
	"email_account_error":           emailAccountRoute,
	// LinkedIn.
	"linkedin_connection_request_sent":     contactRoute,
	"linkedin_connection_request_accepted": contactRoute,
	"linkedin_message_sent":                contactRoute,
	"linkedin_message_replied":             contactRoute,
	"linkedin_reply_categorized":           contactRoute,
	"linkedin_account_alerts":              linkedInAccountRoute,
	"linkedin_comment_posted":              contactRoute,
	"linkedin_comment_skipped":             contactRoute,
	// Calls.
	"contact_called": contactRoute,
	// Contact lifecycle.
	"contact_replied":   contactRoute,
	"contact_finished":  contactRoute,
	"contact_opted_out": contactRoute,
	// Sequence automation.
	"autopilot_stopped": sequenceRoute,
}

// PreLoadData is a no-op for Reply: the payload is self-contained and no header is needed.
func (e SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

// RawMap returns a defensive clone so callers don't mutate the original.
func (e SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// RawMap returns a defensive clone so callers don't mutate the original.
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList returns the payload as a single event, as Reply sends one per delivery.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

// RawEventName returns Reply's event.type verbatim (e.g. "email_replied").
func (e SubscriptionEvent) RawEventName() (string, error) {
	envelope, ok := e[eventFieldEnvelope].(map[string]any)
	if !ok {
		return "", errMissingEventType
	}

	eventType, ok := envelope[eventFieldType].(string)
	if !ok || eventType == "" {
		return "", errMissingEventType
	}

	return eventType, nil
}

// EventType is always "other": Reply reports activity, not record creation, update or deletion.
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	if _, err := e.route(); err != nil {
		return common.SubscriptionEventTypeOther, err
	}

	return common.SubscriptionEventTypeOther, nil
}

// ObjectName returns the object the event concerns (contacts, sequences, ...).
func (e SubscriptionEvent) ObjectName() (string, error) {
	route, err := e.route()
	if err != nil {
		return "", err
	}

	return route.object, nil
}

// Workspace returns an empty string because Reply subscriptions are routed by their URL: each
// installation registers its own webhook URL, so the receiver matches a delivery to its
// installation by that URL rather than by a workspace key.
func (e SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

// RecordId returns the id of the record the event concerns, in the form the by-id endpoint of
// ObjectName takes: contact_fields.id, email_account_id, linkedin_id or sequence_fields.id.
func (e SubscriptionEvent) RecordId() (string, error) {
	route, err := e.route()
	if err != nil {
		return "", err
	}

	var node any = map[string]any(e)

	for _, key := range route.idPath {
		object, ok := node.(map[string]any)
		if !ok {
			return "", errMissingRecordID
		}

		node = object[key]
	}

	return recordIDString(node)
}

// EventTimeStampNano returns the time Reply says the event occurred (event.date).
func (e SubscriptionEvent) EventTimeStampNano() (int64, error) {
	envelope, ok := e[eventFieldEnvelope].(map[string]any)
	if !ok {
		return 0, errMissingEventDate
	}

	date, ok := envelope[eventFieldDate].(string)
	if !ok || date == "" {
		return 0, errMissingEventDate
	}

	occurred, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errUnparsableEventDate, err)
	}

	return occurred.UnixNano(), nil
}

func (e SubscriptionEvent) route() (eventRoute, error) {
	eventType, err := e.RawEventName()
	if err != nil {
		return eventRoute{}, err
	}

	route, ok := eventRoutes[eventType]
	if !ok {
		return eventRoute{}, fmt.Errorf("%w: %s", errUnsupportedEvent, eventType)
	}

	return route, nil
}

// recordIDString renders an id the way Reply's URLs take it. JSON numbers decode to float64.
func recordIDString(value any) (string, error) {
	switch typed := value.(type) {
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case string:
		if typed != "" {
			return typed, nil
		}
	}

	return "", errMissingRecordID
}

// VerifyWebhookMessage always returns true. Reply documents no webhook signature or secret,
// and a live delivery carries no signing header (verified empirically: only content-type
// and traceparent), so there is nothing to verify against. The subscription URL is unique
// per installation.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	return true, nil
}
