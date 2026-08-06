package subscribe

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// mockHubspotEventsPayload is an array-shaped webhook body built from the example events
// documented in providers/hubspot/subscriptionEvent.go (HubSpot delivers a JSON array of
// events per webhook).
const mockHubspotEventsPayload = `[
  {
    "appId": 4210286,
    "eventId": 100,
    "subscriptionId": 2881778,
    "portalId": 44237313,
    "occurredAt": 1731612159499,
    "subscriptionType": "contact.creation",
    "attemptNumber": 0,
    "objectId": 123,
    "changeSource": "CRM",
    "changeFlag": "NEW"
  },
  {
    "appId": 4210286,
    "eventId": 100,
    "subscriptionId": 2902227,
    "portalId": 44237313,
    "occurredAt": 1731612210994,
    "subscriptionType": "contact.propertyChange",
    "attemptNumber": 0,
    "objectId": 123,
    "changeSource": "CRM",
    "propertyName": "message",
    "propertyValue": "sample-value"
  }
]`

// TestMockHubspotConfigResolutionAndEventCasting verifies the mockhubspot registry entry
// resolves like HubSpot's (verification bypassed, look-up-only) and that its event caster is
// HubSpot's real one: an array-shaped webhook body must cast into hubspot.SubscriptionEvent
// values that classify identically to real HubSpot events.
func TestMockHubspotConfigResolutionAndEventCasting(t *testing.T) { //nolint:paralleltest,funlen // mutates the provider catalog
	providers.SetupMockHubspotProvider()

	info, err := providers.ReadInfo(providers.MockHubspot)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	cfg, err := GetProviderConfig("", ResolveProviderInfoAlias(info), deps.Dependencies{})
	if err != nil {
		t.Fatalf("GetProviderConfig: %v", err)
	}

	if !cfg.Verification.ShouldBypass() {
		t.Error("expected mockhubspot webhook verification to be bypassed")
	}

	if cfg.Subscription.IsSupportedViaAPI() {
		t.Error("expected mockhubspot to mirror HubSpot's look-up-only subscribe support")
	}

	var list []map[string]any
	if err := json.Unmarshal([]byte(mockHubspotEventsPayload), &list); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	events, err := cfg.Verification.CastEvents(list)
	if err != nil {
		t.Fatalf("CastEvents: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	created, updated := events[0], events[1]

	if evtType, err := created.EventType(); err != nil || evtType != common.SubscriptionEventTypeCreate {
		t.Errorf("expected create event, got %q (err %v)", evtType, err)
	}

	if objectName, err := created.ObjectName(); err != nil || objectName != "contact" {
		t.Errorf("expected object name %q, got %q (err %v)", "contact", objectName, err)
	}

	if workspace, err := created.Workspace(); err != nil || workspace != "44237313" {
		t.Errorf("expected workspace (portalId) %q, got %q (err %v)", "44237313", workspace, err)
	}

	if recordID, err := created.RecordId(); err != nil || recordID != "123" {
		t.Errorf("expected record id %q, got %q (err %v)", "123", recordID, err)
	}

	if evtType, err := updated.EventType(); err != nil || evtType != common.SubscriptionEventTypeUpdate {
		t.Errorf("expected update event, got %q (err %v)", evtType, err)
	}

	updateEvent, ok := updated.(common.SubscriptionUpdateEvent)
	if !ok {
		t.Fatal("expected cast event to implement SubscriptionUpdateEvent")
	}

	fields, err := updateEvent.UpdatedFields()
	if err != nil {
		t.Fatalf("UpdatedFields: %v", err)
	}

	if len(fields) != 1 || fields[0] != "message" {
		t.Errorf("expected updated fields [message], got %v", fields)
	}
}
