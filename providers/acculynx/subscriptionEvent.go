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

var (
	_ common.SubscriptionEvent       = SubscriptionEvent{}
	_ common.SubscriptionUpdateEvent = SubscriptionEvent{}
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
	errMissingTopicName      = errors.New("acculynx event: missing topicName")
	errUnsupportedTopicName  = errors.New("acculynx event: topicName does not map to a supported object")
	errMissingInnerEvent     = errors.New("acculynx event: missing inner event payload")
	errMissingObjectWrapper  = errors.New("acculynx event: missing object wrapper inside event")
	errMissingSubscriptionID = errors.New("acculynx event: missing subscriptionId")
	errMissingRecordID       = errors.New("acculynx event: missing record id for topic")
	errUnparsableEventTime   = errors.New("acculynx event: unparsable eventDateTime")
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

// PreLoadData is a no-op for AccuLynx — webhook payloads are self-contained.
func (e SubscriptionEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

// RawMap returns a defensive clone so callers don't mutate the original.
func (e SubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// RawEventName returns AccuLynx's topicName verbatim (e.g. "job_created").
func (e SubscriptionEvent) RawEventName() (string, error) {
	topic, ok := e[eventFieldTopicName].(string)
	if !ok || topic == "" {
		return "", errMissingTopicName
	}

	return topic, nil
}

// EventType maps the topic suffix to Create / Update / Other. AccuLynx has
// no delete topics.
func (e SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	topic, err := e.RawEventName()
	if err != nil {
		return common.SubscriptionEventTypeOther, err
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

// ObjectName derives the target object from the topic prefix
// (contact* → contacts, job* → jobs).
func (e SubscriptionEvent) ObjectName() (string, error) {
	topic, err := e.RawEventName()
	if err != nil {
		return "", err
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

// Workspace returns the top-level subscriptionId. AccuLynx enforces one
// subscription per installation, making it a 1:1 proxy for the installation.
func (e SubscriptionEvent) Workspace() (string, error) {
	subID, ok := e[eventFieldSubscriptionID].(string)
	if !ok || subID == "" {
		return "", errMissingSubscriptionID
	}

	return subID, nil
}

// RecordId returns the affected contact/job id from event.<object>.id.
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
// empty slice in UpdatedFields.
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
	"job.representatives.company_assigned":              {"companyRepresentative"},
	"job.representatives.company_changed":               {"companyRepresentative"},
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
	case objectJobs:
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
