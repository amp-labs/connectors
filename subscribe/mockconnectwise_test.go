package subscribe

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// mockConnectWiseWebhookPayload is an object-shaped ConnectWise callback body, a trimmed copy
// of providers/connectwise/internal/webhook/test/contact-update.json. The changed record
// travels inline as an escaped JSON string under Entity.
const mockConnectWiseWebhookPayload = `{
  "Action": "updated",
  "CallbackObjectRecId": 26604,
  "CompanyId": "Cobalt",
  "Entity": "{\"id\":57961,\"firstName\":\"Xzavier Sawayn\",\"inactiveFlag\":false}",
  "FromUrl": "sandbox-na.myconnectwise.net",
  "ID": 57961,
  "MemberId": "Aurasell",
  "MessageId": "6dd82302-5e46-47c0-8c84-702868125b76",
  "Type": "contact"
}`

// TestMockConnectWiseConfigResolution verifies the mockconnectwise registry entry resolves
// like ConnectWise's: verification bypassed, the WatchFieldsAuto quirk, and subscribe-by-API
// support mirrored.
func TestMockConnectWiseConfigResolution(t *testing.T) { //nolint:paralleltest // mutates the provider catalog
	providers.SetupMockConnectWiseProvider()

	info, err := providers.ReadInfo(providers.MockConnectWise)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	cfg, err := GetProviderConfig("", ResolveProviderInfoAlias(info), deps.Dependencies{})
	if err != nil {
		t.Fatalf("GetProviderConfig: %v", err)
	}

	if !cfg.Verification.ShouldBypass() {
		t.Error("expected mockconnectwise webhook verification to be bypassed")
	}

	if !cfg.Subscription.RequiresWatchFieldsAutoAll() {
		t.Error("expected mockconnectwise to mirror ConnectWise's WatchFieldsAuto quirk")
	}

	if !cfg.Subscription.IsSupportedViaAPI() {
		t.Error("expected mockconnectwise to mirror ConnectWise's subscribe-by-API support")
	}
}

// TestMockConnectWiseEventParsing runs a ConnectWise-shaped callback body through the same
// object-payload dispatch the server's receive path uses and asserts ConnectWise's real event
// implementation classifies it — including the inline-record Entity extraction
// (SubscriptionEventWithRecord) the receive path falls back to when a record is not queryable.
func TestMockConnectWiseEventParsing(t *testing.T) { //nolint:funlen
	t.Parallel()

	var rawEvent map[string]any
	if err := json.Unmarshal([]byte(mockConnectWiseWebhookPayload), &rawEvent); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	events, err := GetObjectTypeSubscribeEventsList(providers.MockConnectWise, rawEvent)
	if err != nil {
		t.Fatalf("GetObjectTypeSubscribeEventsList: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected ConnectWise's one-record-per-callback shape, got %d events", len(events))
	}

	event := events[0]

	if evtType, err := event.EventType(); err != nil || evtType != common.SubscriptionEventTypeUpdate {
		t.Errorf("expected update event, got %q (err %v)", evtType, err)
	}

	if objectName, err := event.ObjectName(); err != nil || objectName != "contacts" {
		t.Errorf("expected object name %q, got %q (err %v)", "contacts", objectName, err)
	}

	if recordID, err := event.RecordId(); err != nil || recordID != "57961" {
		t.Errorf("expected record id %q, got %q (err %v)", "57961", recordID, err)
	}

	withRecord, ok := event.(common.SubscriptionEventWithRecord)
	if !ok {
		t.Fatal("expected ConnectWise event to carry its record inline (SubscriptionEventWithRecord)")
	}

	row, err := withRecord.Record([]string{"firstName"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	if row.Id != "57961" {
		t.Errorf("expected inline record id %q, got %q", "57961", row.Id)
	}

	if got := row.Fields["firstname"]; got != "Xzavier Sawayn" {
		t.Errorf("expected requested field in Fields, got %v", got)
	}

	if got := row.Raw["inactiveFlag"]; got != false {
		t.Errorf("expected Raw to carry the full inline record, got inactiveFlag=%v", got)
	}
}
