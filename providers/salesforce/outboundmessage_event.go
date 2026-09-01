package salesforce

import (
	"encoding/xml"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

// This file parses the SOAP messages Salesforce POSTs to the outbound message
// endpoint of a flow-based subscription (see subscribe_flow.go).
//
// Wire format: one SOAP 1.1 envelope whose body carries a <notifications>
// element (namespace http://soap.sforce.com/2005/09/outbound) with envelope-
// level metadata (OrganizationId, ActionId, EnterpriseUrl, PartnerUrl) and up
// to 100 <Notification> children. Each Notification holds the message id and
// an <sObject> whose xsi:type names the object (e.g. "sf:Account") and whose
// children are the fields configured on the outbound message.
// https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_om_outboundmessaging_understanding.htm
//
// The receiver must respond with an Ack=true SOAP body (BuildOutboundMessageAck)
// or Salesforce keeps retrying delivery for up to 24 hours.
//
// Outbound messages carry NO event type. The parser infers it from the audit
// fields the outbound message builder always includes (see
// ensureOutboundMessageFields): CreatedDate == LastModifiedDate means the save
// that fired the flow created the record; a later LastModifiedDate means an
// update. Deletes never produce outbound messages.
//
// Outbound messages are also unsigned — there is no HMAC. Verification options
// are the unguessable per-subscription endpoint URL (current posture, matching
// VerifyWebhookMessage's allow-all), validating Salesforce's client certificate
// at the TLS layer, or checking OrganizationId against the connection.

var (
	errNotOutboundMessage = errors.New("body is not a Salesforce outbound message notification")
	errMissingSObject     = errors.New("outbound message notification carries no sObject")
	errMissingTimestamps  = errors.New("outbound message sObject carries no CreatedDate/LastModifiedDate")
)

const (
	omKeyObjectName     = "objectName"
	omKeyNotificationID = "notificationId"
	omKeyOrganizationID = "organizationId"
	omKeyActionID       = "actionId"
	omKeyEnterpriseURL  = "enterpriseUrl"
	omKeySObject        = "sObject"

	omFieldID               = "Id"
	omFieldCreatedDate      = "CreatedDate"
	omFieldLastModifiedDate = "LastModifiedDate"

	rawEventNameCreate          = "CREATE"
	rawEventNameUpdate          = "UPDATE"
	rawEventNameOutboundMessage = "OUTBOUND_MESSAGE"
)

// soap/xml shapes for the inbound notification envelope. They mirror the
// notifications element of the outbound messaging WSDL; official docs with an
// example SOAP message and the required acknowledgment:
// https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_om_outboundmessaging_understanding.htm
// https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_om_outboundmessaging_wsdl.htm

type omEnvelopeXML struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Notifications omNotificationsXML `xml:"notifications"`
	} `xml:"Body"`
}

type omNotificationsXML struct {
	OrganizationID string              `xml:"OrganizationId"`
	ActionID       string              `xml:"ActionId"`
	EnterpriseURL  string              `xml:"EnterpriseUrl"`
	PartnerURL     string              `xml:"PartnerUrl"`
	Notifications  []omNotificationXML `xml:"Notification"`
}

type omNotificationXML struct {
	ID      string       `xml:"Id"`
	SObject omSObjectXML `xml:"sObject"`
}

type omSObjectXML struct {
	// Type mirrors the xsi:type attribute, e.g. "sf:Account".
	Type   string       `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	Fields []omFieldXML `xml:",any"`
}

type omFieldXML struct {
	XMLName xml.Name
	Nil     string `xml:"http://www.w3.org/2001/XMLSchema-instance nil,attr"`
	Value   string `xml:",chardata"`
}

// ParseOutboundMessage parses the SOAP body of one outbound message POST into
// a collapsed event that expands to one SubscriptionEvent per Notification.
func ParseOutboundMessage(body []byte) (*OutboundMessageEnvelope, error) {
	var envelope omEnvelopeXML

	// nolint:musttag // omFieldXML.XMLName is intentionally untagged: it captures
	// each sObject child element's dynamic name (the configured field names).
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %w", errNotOutboundMessage, err)
	}

	notifications := envelope.Body.Notifications
	if len(notifications.Notifications) == 0 {
		return nil, fmt.Errorf("%w: no Notification elements", errNotOutboundMessage)
	}

	events := make([]OutboundMessageEvent, 0, len(notifications.Notifications))

	for _, notification := range notifications.Notifications {
		if len(notification.SObject.Fields) == 0 && notification.SObject.Type == "" {
			return nil, fmt.Errorf("%w: notification %s", errMissingSObject, notification.ID)
		}

		fields := make(map[string]any, len(notification.SObject.Fields))

		for _, field := range notification.SObject.Fields {
			if field.Nil == "true" {
				fields[field.XMLName.Local] = nil

				continue
			}

			fields[field.XMLName.Local] = field.Value
		}

		events = append(events, OutboundMessageEvent{
			omKeyNotificationID: notification.ID,
			omKeyObjectName:     objectNameFromSObjectType(notification.SObject.Type),
			omKeyOrganizationID: notifications.OrganizationID,
			omKeyActionID:       notifications.ActionID,
			omKeyEnterpriseURL:  notifications.EnterpriseURL,
			omKeySObject:        fields,
		})
	}

	return &OutboundMessageEnvelope{events: events}, nil
}

// objectNameFromSObjectType strips the namespace prefix from an xsi:type value:
// "sf:Account" → "Account".
func objectNameFromSObjectType(sObjectType string) string {
	if _, name, found := strings.Cut(sObjectType, ":"); found {
		return name
	}

	return sObjectType
}

var _ common.CollapsedSubscriptionEvent = &OutboundMessageEnvelope{}

// OutboundMessageEnvelope is the collapsed form of one outbound message POST,
// which batches up to 100 notifications.
type OutboundMessageEnvelope struct {
	events []OutboundMessageEvent
}

func (e *OutboundMessageEnvelope) RawMap() (map[string]any, error) {
	raw := make([]any, len(e.events))
	for index, event := range e.events {
		raw[index] = map[string]any(event)
	}

	return map[string]any{"notifications": raw}, nil
}

func (e *OutboundMessageEnvelope) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	events := make([]common.SubscriptionEvent, len(e.events))
	for index, event := range e.events {
		events[index] = event
	}

	return events, nil
}

var _ common.SubscriptionEvent = OutboundMessageEvent{}

// OutboundMessageEvent is one Notification from an outbound message, flattened
// into a map with envelope-level metadata attached.
type OutboundMessageEvent map[string]any

func (e OutboundMessageEvent) PreLoadData(_ *common.SubscriptionEventPreLoadData) error {
	return nil
}

func (e OutboundMessageEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

// EventType infers create vs update from the record's audit timestamps, which
// the outbound message builder always includes: a record whose
// LastModifiedDate equals its CreatedDate was created by the save that fired
// the flow. When either timestamp is missing (caller-managed outbound message
// without audit fields), the type is Other rather than a guess.
func (e OutboundMessageEvent) EventType() (common.SubscriptionEventType, error) {
	created, createdErr := e.fieldTime(omFieldCreatedDate)
	modified, modifiedErr := e.fieldTime(omFieldLastModifiedDate)

	if createdErr != nil || modifiedErr != nil {
		return common.SubscriptionEventTypeOther, nil //nolint:nilerr
	}

	if created.Equal(modified) {
		return common.SubscriptionEventTypeCreate, nil
	}

	return common.SubscriptionEventTypeUpdate, nil
}

func (e OutboundMessageEvent) RawEventName() (string, error) {
	eventType, err := e.EventType()
	if err != nil {
		return "", err
	}

	switch eventType { //nolint:exhaustive
	case common.SubscriptionEventTypeCreate:
		return rawEventNameCreate, nil
	case common.SubscriptionEventTypeUpdate:
		return rawEventNameUpdate, nil
	default:
		return rawEventNameOutboundMessage, nil
	}
}

func (e OutboundMessageEvent) ObjectName() (string, error) {
	return common.StringMap(e).GetString(omKeyObjectName)
}

func (e OutboundMessageEvent) Workspace() (string, error) {
	// Not applicable: the endpoint URL is unique per subscription, so the
	// receiver already knows which connection the message belongs to.
	return "", nil
}

func (e OutboundMessageEvent) RecordId() (string, error) {
	fields, err := e.sObjectFields()
	if err != nil {
		return "", err
	}

	return fields.GetString(omFieldID)
}

// EventTimeStampNano reports when the change was committed, using the
// record's LastModifiedDate (CreatedDate as fallback).
func (e OutboundMessageEvent) EventTimeStampNano() (int64, error) {
	if modified, err := e.fieldTime(omFieldLastModifiedDate); err == nil {
		return modified.UnixNano(), nil
	}

	created, err := e.fieldTime(omFieldCreatedDate)
	if err != nil {
		return 0, errMissingTimestamps
	}

	return created.UnixNano(), nil
}

func (e OutboundMessageEvent) sObjectFields() (common.StringMap, error) {
	fieldsAny, ok := e[omKeySObject]
	if !ok {
		return nil, errMissingSObject
	}

	fields, ok := fieldsAny.(map[string]any)
	if !ok {
		return nil, errMissingSObject
	}

	return common.StringMap(fields), nil
}

func (e OutboundMessageEvent) fieldTime(fieldName string) (time.Time, error) {
	fields, err := e.sObjectFields()
	if err != nil {
		return time.Time{}, err
	}

	value, err := fields.GetString(fieldName)
	if err != nil {
		return time.Time{}, err
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse %s %q: %w", fieldName, value, err)
	}

	return parsed, nil
}

// outboundMessageAckTemplate is the SOAP body Salesforce expects in response
// to an outbound message POST. Anything other than Ack=true (or a non-2xx
// status) makes Salesforce retry delivery for up to 24 hours.
const outboundMessageAckTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
    <soapenv:Body>
        <notificationsResponse xmlns="http://soap.sforce.com/2005/09/outbound">
            <Ack>%t</Ack>
        </notificationsResponse>
    </soapenv:Body>
</soapenv:Envelope>
`

// BuildOutboundMessageAck returns the SOAP acknowledgment body the receiver
// must send back for an outbound message POST. Pass false to make Salesforce
// redeliver the batch later (e.g. on transient downstream failure).
func BuildOutboundMessageAck(ack bool) []byte {
	return fmt.Appendf(nil, outboundMessageAckTemplate, ack)
}
