// nolint:lll,tagliatelle
package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (*Connector) VerifyWebhookMessage(context.Context, *common.WebhookVerificationParameters) (bool, error) {
	return true, nil
}

// ChangeEvent represents data received from a subscription.
// A single event may contain multiple record identifiers and can be expanded into multiple SubscriptionEvent instances.
//
// Structure reference:
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_event_fields_header.htm.
type ChangeEvent struct {
	ChangeEventHeader ChangeEventHeader `json:"ChangeEventHeader"`
	// ...
	// Additional fields represent changes to the record and vary based on the record's model type.
	// They are situated on the same level as ChangeEventHeader.
	// ...
}

// ChangeEventHeader holds event data.
//
// Record ID can have wildcard symbols:
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_field_conversion_single_event.htm.
type ChangeEventHeader struct {
	EntityName      string   `json:"entityName"`
	RecordIDs       []string `json:"recordIds"`
	ChangeType      string   `json:"changeType"`
	ChangeOrigin    string   `json:"changeOrigin"`
	TransactionKey  string   `json:"transactionKey"`
	SequenceNumber  int      `json:"sequenceNumber"`
	CommitTimestamp int64    `json:"commitTimestamp"`
	CommitNumber    int64    `json:"commitNumber"`
	CommitUser      string   `json:"commitUser"`
	NulledFields    []any    `json:"nulledFields"`
	DiffFields      []any    `json:"diffFields"`
	ChangedFields   []any    `json:"changedFields"`
}

// Unwrap splits bundled event into per record event.
func (e ChangeEvent) Unwrap() []SubscriptionEvent {
	header := e.ChangeEventHeader
	events := make([]SubscriptionEvent, len(header.RecordIDs))

	for index, recordID := range header.RecordIDs {
		events[index] = SubscriptionEvent{
			EntityName:       header.EntityName,
			RecordIdentifier: recordID,
			ChangeType:       header.ChangeType,
			ChangeOrigin:     header.ChangeOrigin,
			TransactionKey:   header.TransactionKey,
			SequenceNumber:   header.SequenceNumber,
			CommitTimestamp:  header.CommitTimestamp,
			CommitNumber:     header.CommitNumber,
			CommitUser:       header.CommitUser,
			NulledFields:     header.NulledFields,
			DiffFields:       header.DiffFields,
			ChangedFields:    header.ChangedFields,
		}
	}

	return events
}

var _ common.SubscriptionEvent = SubscriptionEvent{}

type SubscriptionEvent struct {
	EntityName       string
	RecordIdentifier string
	ChangeType       string
	ChangeOrigin     string
	TransactionKey   string
	SequenceNumber   int
	CommitTimestamp  int64
	CommitNumber     int64
	CommitUser       string
	NulledFields     []any
	DiffFields       []any
	ChangedFields    []any
}

func (s SubscriptionEvent) EventType() (common.SubscriptionEventType, error) {
	switch s.ChangeType {
	case "CREATE", "GAP_CREATE":
		return common.SubscriptionEventTypeCreate, nil
	case "UPDATE", "GAP_UPDATE":
		return common.SubscriptionEventTypeUpdate, nil
	case "DELETE", "GAP_DELETE":
		return common.SubscriptionEventTypeDelete, nil
	case "UNDELETE", "SNAPSHOT", "GAP_UNDELETE", "GAP_OVERFLOW":
		fallthrough
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (s SubscriptionEvent) RawEventName() (string, error) {
	return s.ChangeType, nil
}

func (s SubscriptionEvent) ObjectName() (string, error) {
	// The API name of the standard or custom object that the change pertains to.
	// For example, Account or MyObject__c.
	return s.EntityName, nil
}

func (s SubscriptionEvent) Workspace() (string, error) {
	// Not applicable
	return "", nil
}

// RecordId
// https://developer.salesforce.com/docs/atlas.en-us.change_data_capture.meta/change_data_capture/cdc_field_conversion_single_event.htm.
func (s SubscriptionEvent) RecordId() (string, error) {
	return s.RecordIdentifier, nil
}

func (s SubscriptionEvent) EventTimeStampNano() (int64, error) {
	// The date and time when the change occurred,
	// represented as the number of milliseconds since January 1, 1970 00:00:00 GMT.
	return s.CommitTimestamp, nil
}
