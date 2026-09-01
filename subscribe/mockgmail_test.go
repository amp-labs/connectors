package subscribe

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// mockGmailEventsPayload is an array of the synthetic Gmail events the Gmail event workflow
// republishes into the subscribe pipeline (google.SubscriptionEvent's documented wire shape:
// one event per affected message, derived from a history.list expansion — Gmail's raw push
// carries only {emailAddress, historyId} and never reaches this path).
const mockGmailEventsPayload = `[
  {
    "messageId": "18c2f4a9b7d3e510",
    "historyId": "5723461",
    "emailAddress": "consumer@example.com",
    "rawEventName": "messagesAdded",
    "occurredAt": 1731612159499000000
  },
  {
    "messageId": "18c2f4a9b7d3e511",
    "historyId": "5723462",
    "emailAddress": "consumer@example.com",
    "rawEventName": "labelsAdded",
    "occurredAt": 1731612210994000000
  },
  {
    "messageId": "18c2f4a9b7d3e512",
    "historyId": "5723463",
    "emailAddress": "consumer@example.com",
    "rawEventName": "messagesDeleted",
    "occurredAt": 1731612299001000000
  }
]`

// TestMockGmailConfigResolutionAndEventCasting verifies the mockgmail registry entry resolves
// like the Gmail module's (verification bypassed, subscribe-by-API with maintenance) and that
// its event caster is Gmail's real one: synthetic republished payloads must cast into
// google.SubscriptionEvent values that classify identically to real Gmail events.
func TestMockGmailConfigResolutionAndEventCasting(t *testing.T) { //nolint:paralleltest,funlen // mutates the provider catalog
	providers.SetupMockGmailProvider()

	info, err := providers.ReadInfo(providers.MockGmail)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	cfg, err := GetProviderConfig("", ResolveProviderInfoAlias(info), deps.Dependencies{})
	if err != nil {
		t.Fatalf("GetProviderConfig: %v", err)
	}

	if !cfg.Verification.ShouldBypass() {
		t.Error("expected mockgmail webhook verification to be bypassed")
	}

	if !cfg.Subscription.IsSupportedViaAPI() {
		t.Error("expected mockgmail to mirror Gmail's subscribe-by-API support")
	}

	var list []map[string]any
	if err := json.Unmarshal([]byte(mockGmailEventsPayload), &list); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	events, err := cfg.Verification.CastEvents(list)
	if err != nil {
		t.Fatalf("CastEvents: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	wantByIndex := []struct {
		eventType common.SubscriptionEventType
		recordID  string
	}{
		{common.SubscriptionEventTypeCreate, "18c2f4a9b7d3e510"},
		{common.SubscriptionEventTypeUpdate, "18c2f4a9b7d3e511"},
		{common.SubscriptionEventTypeDelete, "18c2f4a9b7d3e512"},
	}

	for i, event := range events {
		if evtType, err := event.EventType(); err != nil || evtType != wantByIndex[i].eventType {
			t.Errorf("event %d: expected %q event, got %q (err %v)", i, wantByIndex[i].eventType, evtType, err)
		}

		if recordID, err := event.RecordId(); err != nil || recordID != wantByIndex[i].recordID {
			t.Errorf("event %d: expected record id %q, got %q (err %v)", i, wantByIndex[i].recordID, recordID, err)
		}

		if objectName, err := event.ObjectName(); err != nil || objectName != "messages" {
			t.Errorf("event %d: expected object name %q, got %q (err %v)", i, "messages", objectName, err)
		}

		if workspace, err := event.Workspace(); err != nil || workspace != "consumer@example.com" {
			t.Errorf("event %d: expected workspace (mailbox) %q, got %q (err %v)", i, "consumer@example.com", workspace, err)
		}
	}
}
