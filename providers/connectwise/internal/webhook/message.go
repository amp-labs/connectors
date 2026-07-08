package webhook

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors/common"
)

// CollapsedSubscriptionEvent represents the raw webhook payload.
// This simply wraps the single event.
type CollapsedSubscriptionEvent map[string]any

// Event is a singular notification message within EventCollection.
type Event map[string]any

var (
	_ common.SubscriptionEvent           = Event{}
	_ common.SubscriptionUpdateEvent     = Event{}
	_ common.SubscriptionEventWithRecord = Event{}
	_ common.CollapsedSubscriptionEvent  = CollapsedSubscriptionEvent{}
)

func (e CollapsedSubscriptionEvent) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e CollapsedSubscriptionEvent) SubscriptionEventList() ([]common.SubscriptionEvent, error) {
	return []common.SubscriptionEvent{Event(e)}, nil
}

func (e Event) EventType() (common.SubscriptionEventType, error) {
	actionStr, err := e.RawEventName()
	if err != nil {
		return "", err
	}

	switch actionStr {
	case "added":
		return common.SubscriptionEventTypeCreate, nil
	case "updated":
		return common.SubscriptionEventTypeUpdate, nil
	case "deleted":
		return common.SubscriptionEventTypeDelete, nil
	default:
		return common.SubscriptionEventTypeOther, nil
	}
}

func (e Event) RawEventName() (string, error) {
	actionStr, ok := e["Action"].(string)
	if !ok {
		return "", fmt.Errorf("%w: 'Action'", common.ErrMissingField)
	}

	return actionStr, nil
}

func (e Event) ObjectName() (string, error) {
	objectType, ok := e["Type"].(string)
	if !ok {
		return "", fmt.Errorf("%w: 'Type'", common.ErrMissingField)
	}

	objectName, ok := ObjectTypeToObjectName[objectType]
	if !ok {
		return "", fmt.Errorf("object 'Type' is unknown (value=%v), cannot map to objectName", objectType) // nolint:err113
	}

	return objectName.String(), nil
}

func (e Event) Workspace() (string, error) {
	return "", nil
}

func (e Event) RecordId() (string, error) {
	identifier, ok := e["ID"].(float64)
	if !ok {
		return "", fmt.Errorf("%w: 'ID'", common.ErrMissingField)
	}

	return fmt.Sprintf("%v", identifier), nil
}

func (e Event) EventTimeStampNano() (int64, error) {
	return 0, nil
}

func (e Event) RawMap() (map[string]any, error) {
	return maps.Clone(e), nil
}

func (e Event) PreLoadData(data *common.SubscriptionEventPreLoadData) error {
	return nil
}

func (e Event) UpdatedFields() ([]string, error) {
	return nil, nil
}

// Record returns the full record carried inline in the webhook payload.
// ConnectWise embeds the changed record as an escaped JSON string under the
// "Entity" field, so we unmarshal it into the same shape a read returns.
func (e Event) Record() (map[string]any, error) {
	entity, ok := e["Entity"].(string)
	if !ok || entity == "" {
		return nil, fmt.Errorf("%w: 'Entity'", common.ErrMissingField)
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(entity), &record); err != nil {
		return nil, fmt.Errorf("parsing connectwise webhook 'Entity': %w", err)
	}

	return record, nil
}

// ObjectNameToObjectType maps ConnectWise connector object names
// to their corresponding callback/webhook object type names.
var ObjectNameToObjectType = map[common.ObjectName]string{ // nolint:gochecknoglobals
	"activities":      "activity",
	"agreements":      "agreement",
	"catalog":         "productcatalog",
	"companies":       "company",
	"configurations":  "configuration",
	"contacts":        "contact",
	"expense/entries": "expense",
	"invoices":        "invoice",
	// "service/tickets" also maps to Ticket.
	// To avoid complex mapping only one ObjectName will be accepted by the Subscribe action.
	"project/tickets":     "ticket",
	"projects":            "project",
	"purchaseorders":      "purchaseorder",
	"sales/opportunities": "opportunity",
	"schedule/entries":    "schedule",
	// Not supported:
	// Site:	"/company/companies/{parentId}/sites"
	"system/members": "member",
	"time/entries":   "time",
}

// ObjectTypeToObjectName is the reverse mapping of ObjectNameToObjectType.
var ObjectTypeToObjectName = map[string]common.ObjectName{ // nolint:gochecknoglobals
	"activity":       "activities",
	"agreement":      "agreements",
	"productcatalog": "catalog",
	"company":        "companies",
	"configuration":  "configurations",
	"contact":        "contacts",
	"expense":        "expense/entries",
	"invoice":        "invoices",
	"ticket":         "project/tickets",
	"project":        "projects",
	"purchaseorder":  "purchaseorders",
	"opportunity":    "sales/opportunities",
	"schedule":       "schedule/entries",
	"member":         "system/members",
	"time":           "time/entries",
}
