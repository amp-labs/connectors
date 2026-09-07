package acculynx

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// SubscriptionEvent is the parsed AccuLynx webhook payload. The documented
// payload shape is stale: real deliveries use lowercase "event", nest the
// record under event.<object>.id, and omit companyId entirely.
//
// Sample delivery (contact_added):
//
//	{
//	  "topicName":      "contact_added",
//	  "eventDateTime":  "2026-05-26T14:23:55.4999782Z",
//	  "eventId":        "38e4c045-2a6c-43f2-8309-ac8b5fc3fc2b",
//	  "subscriptionId": "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8",
//	  "event": {
//	    "contact": {
//	      "id":   "eadaaa11-1276-4166-bb93-db02f46b39a2",
//	      "date": "2026-05-26T14:23:55.2976156Z",
//	      "_link": "https://api.acculynx.com/api/v2/contacts/eadaaa11-..."
//	    }
//	  }
//	}
type SubscriptionEvent map[string]any

// CollapsedSubscriptionEvent is the raw AccuLynx webhook payload as received.
// AccuLynx delivers exactly one event per request (a single topicName and a
// single event.<object> wrapper), so the collapsed form expands to a one-element
// list. This is the entry point the webhook processor uses to extract the
// individual SubscriptionEvent(s) from a delivery; without it the processor
// cannot turn an accepted webhook into a deliverable event. Mirrors the
// single-event convention used by housecallPro/outreach.
type CollapsedSubscriptionEvent map[string]any

var (
	_ common.SubscriptionEvent          = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent    = SubscriptionEvent{}
	_ common.CollapsedSubscriptionEvent = CollapsedSubscriptionEvent{}
)

const (
	eventFieldTopicName      = "topicName"
	eventFieldEventDateTime  = "eventDateTime"
	eventFieldEvent          = "event"
	eventFieldSubscriptionID = "subscriptionId"

	objectWrapperJob     = "job"
	objectWrapperContact = "contact"

	innerFieldID = "id"
)

var (
	errMissingTopicName     = errors.New("acculynx event: missing topicName")
	errUnsupportedTopicName = errors.New("acculynx event: topicName does not map to a supported object")
	errMissingInnerEvent    = errors.New("acculynx event: missing inner event payload")
	errMissingObjectWrapper = errors.New("acculynx event: missing object wrapper inside event")
	errMissingRecordID      = errors.New("acculynx event: missing record id for topic")
	errUnparsableEventTime  = errors.New("acculynx event: unparsable eventDateTime")
	// errParentRecordIDUnavailable is returned for topics whose payload omits
	// the parent contact/job id entirely (only the changed sub-object is sent).
	errParentRecordIDUnavailable = errors.New(
		"acculynx event: parent record id is not available in payload for this topic")
)

//nolint:gochecknoglobals
var topicsWithoutParentRecordID = datautils.NewStringSet(
	"contact.custom-field.status_changed",
	"job.custom-field.status_changed",
)

// topicObjectOverrides maps topics that belong to a nested subscribe object,
// consulted before the topic-prefix fallback in ObjectName. The company
// representative is a per-job singleton with its own get-by-id path, so its
// topics resolve to jobs/representatives rather than jobs (the inverse of
// housecallPro, which routes employee.* to "other" because no get-by-id
// endpoint exists there).
//
//nolint:gochecknoglobals
var topicObjectOverrides = map[string]string{
	"job.representatives.company_assigned": objectJobsRepresentatives,
	"job.representatives.company_changed":  objectJobsRepresentatives,
}

// topicEventTypeOverrides pins topics whose suffix misclassifies them.
// company_assigned ends in _assigned (suffix heuristic: other), but it is an
// update to the job's representative slot: the slot always exists, assignment
// fills it. Consulted before the suffix switch in EventType.
//
//nolint:gochecknoglobals
var topicEventTypeOverrides = map[string]common.SubscriptionEventType{
	"job.representatives.company_assigned": common.SubscriptionEventTypeUpdate,
}

// PreLoadData is a no-op for AccuLynx — webhook payloads are self-contained.
func (e SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

// RawMap returns a defensive clone so callers don't mutate the original.
func (e SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// RawMap returns a defensive clone so callers don't mutate the original
// (separate from SubscriptionEvent.RawMap because Go method sets are per-type;
// CollapsedSubscriptionEvent must satisfy common.Event in its own right).
func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// SubscriptionEventList expands the collapsed payload into the individual events
// it contains. AccuLynx sends one event per webhook delivery, so this returns a
// single-element list wrapping the payload as a SubscriptionEvent.
func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{SubscriptionEvent(e)}, nil
}

// RawEventName returns AccuLynx's topicName verbatim (e.g. "job_created").
func (e SubscriptionEvent) RawEventName() (string, error) {
	topic, ok := e[eventFieldTopicName].(string)
	if !ok || topic == "" {
		return "", errMissingTopicName
	}

	return topic, nil
}

// EventType classifies the topic: exact overrides first (see
// topicEventTypeOverrides), then the suffix heuristic. AccuLynx has no
// delete topics.
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	topic, err := e.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, err
	}

	if evtType, ok := topicEventTypeOverrides[topic]; ok {
		return evtType, nil
	}

	switch {
	case strings.HasSuffix(topic, "_added"), strings.HasSuffix(topic, "_created"):
		return common.SubscriptionEventTypeCreate, nil
	case strings.HasSuffix(topic, "_changed"),
		strings.HasSuffix(topic, "_updated"),
		strings.HasSuffix(topic, "_voided"):
		return common.SubscriptionEventTypeUpdate, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

// ObjectName resolves the target object: exact topic overrides first (see
// topicObjectOverrides, e.g. jobs/representatives), then the topic prefix
// (contact* → contacts, job* → jobs).
func (e SubscriptionEvent) ObjectName() (string, error) {
	topic, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	if obj, ok := topicObjectOverrides[topic]; ok {
		return obj, nil
	}

	switch {
	case strings.HasPrefix(topic, "contact"):
		return objectContacts, nil
	case strings.HasPrefix(topic, "job"):
		return objectJobs, nil
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedTopicName, topic)
	}
}

// Workspace returns an empty string because AccuLynx subscriptions are
// consumer-URL routed: each installation registers a unique, immutable
// consumerUrl, so the receiver matches a webhook to its installation by that URL
// — not by a workspace key. AccuLynx implements no GetPostAuthInfo, so no
// installation carries a providerWorkspaceRef; returning a non-empty value here
// (e.g. the subscriptionId) leaves the receiver with nothing to match, so it
// accepts the event (HTTP 200) but never routes it to a destination. This
// follows the URL-routed convention used by salesforce/housecallPro/outreach.
func (e SubscriptionEvent) Workspace() (string, error) {
	return "", nil
}

// RecordId returns the affected record id from event.<wrapper>.id. For
// contacts/jobs that is the record's own id; for jobs/representatives it is
// the job id read from the same event.job wrapper — the slot's only stable
// identity, since AccuLynx rotates the representative id on every assignment.
// Returns errParentRecordIDUnavailable for topics where AccuLynx omits the
// parent id (see topicsWithoutParentRecordID).
func (e SubscriptionEvent) RecordId() (string, error) {
	topic, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	if topicsWithoutParentRecordID.Has(topic) {
		return "", fmt.Errorf("%w (%s)", errParentRecordIDUnavailable, topic)
	}

	wrapper, err := e.objectWrapper()
	if err != nil {
		return "", err
	}

	id, ok := wrapper[innerFieldID].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("%w (%s)", errMissingRecordID, topic)
	}

	return id, nil
}

// EventTimeStampNano parses AccuLynx's RFC3339Nano eventDateTime and returns
// nanoseconds since epoch.
func (e SubscriptionEvent) EventTimeStampNano() (int64, error) {
	raw, ok := e[eventFieldEventDateTime].(string)
	if !ok || raw == "" {
		return 0, fmt.Errorf("%w: %v", errUnparsableEventTime, e[eventFieldEventDateTime])
	}

	t, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", errUnparsableEventTime, err)
	}

	return t.UnixNano(), nil
}

// topicToUpdatedFields maps each specific-change topic to its wire-format
// field name under event.<object>. Generic and create topics resolve to an
// empty slice in UpdatedFields. Representative topics are absent on purpose:
// they are updates to the jobs/representatives object itself, where no single
// field can express the change, so UpdatedFields stays empty and consumers
// watch with watchFieldsAuto.
//
//nolint:gochecknoglobals
var topicToUpdatedFields = map[string][]string{
	"job.milestone.current_changed":                     {"milestone"},
	"job.milestone.status.current_changed":              {"milestone"},
	"job.financials.approved-value_changed":             {"financials"},
	"job.category_changed":                              {"jobCategory"},
	"job.work-type_changed":                             {"workType"},
	"job.trade-type_changed":                            {"tradeTypes"},
	"job.contacts.primary_changed":                      {"contacts"},
	"job.appointments.initial_created":                  {"initialAppointment"},
	"job.appointments.initial_updated":                  {"initialAppointment"},
	"job.invoice_updated":                               {"invoice"},
	"job.invoice_voided":                                {"invoice"},
	"job.custom-field.value_changed":                    {"customField"},
	"job.custom-field.status_changed":                   {"customField"},
	"contact.custom-field.value_changed":                {"customField"},
	"contact.custom-field.status_changed":               {"customField"},
	"job.accounting.integration-status.current_changed": {"accounting"},
}

// UpdatedFields returns the field name(s) the topic implies. Empty slice for
// generic update/create topics.
func (e SubscriptionEvent) UpdatedFields() ([]string, error) {
	topic, err := e.RawEventName()
	if err != nil {
		return nil, err
	}

	if fields, ok := topicToUpdatedFields[topic]; ok {
		return fields, nil
	}

	return []string{}, nil
}

func (e SubscriptionEvent) innerEvent() (map[string]any, error) {
	inner, ok := e[eventFieldEvent].(map[string]any)
	if !ok {
		return nil, errMissingInnerEvent
	}

	return inner, nil
}

func (e SubscriptionEvent) objectWrapper() (map[string]any, error) {
	objName, err := e.ObjectName()
	if err != nil {
		return nil, err
	}

	inner, err := e.innerEvent()
	if err != nil {
		return nil, err
	}

	var key string

	switch objName {
	case objectContacts:
		key = objectWrapperContact
	case objectJobs, objectJobsRepresentatives:
		// Representative topics are delivered inside the job wrapper
		// (event.job.companyRepresentative); the record id is the job id.
		key = objectWrapperJob
	}

	wrapper, ok := inner[key].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: expected event.%s", errMissingObjectWrapper, key)
	}

	return wrapper, nil
}

// VerifyWebhookMessage always returns true. AccuLynx's docs reference a
// webhook secret, but the live API does not return one on subscription create
// and deliveries carry no signing header (verified empirically).
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	return true, nil
}
