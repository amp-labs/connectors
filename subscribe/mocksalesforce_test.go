package subscribe

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// mockSalesforceEventBridgePayload is a Salesforce CDC event inside the AWS EventBridge
// envelope the receive path sees (metadata layers around detail.payload). The CDC payload is a
// trimmed copy of providers/salesforce/test/subscription/new_account.json; recordIds carries
// two ids to exercise the per-record fan-out.
const mockSalesforceEventBridgePayload = `{
  "version": "0",
  "id": "5c42b99e-c214-4a9c-8fbd-1a41e5faa44a",
  "detail-type": "AccountChangeEvent",
  "source": "aws.partner/salesforce.com/00D5f000005KcCnEAK/0YL5f000000CabWGAS",
  "detail": {
    "payload": {
      "ChangeEventHeader": {
        "entityName": "Account",
        "recordIds": [
          "0015f00002J9YYEAA3",
          "0015f00002J9YYFAA3"
        ],
        "changeType": "CREATE",
        "changeOrigin": "com/salesforce/api/soap/60.0;client=SfdcInternalAPI/",
        "transactionKey": "0001ade9-3f74-0b99-dbc4-42e73424b774",
        "sequenceNumber": 1,
        "commitTimestamp": 1712693965000,
        "commitUser": "0055f000005mc66AAA",
        "nulledFields": [],
        "diffFields": [],
        "changedFields": []
      },
      "Name": "Acme",
      "Description": "Sample account record.",
      "OwnerId": "0055f000005mc66AAA",
      "CreatedDate": 1712693965000
    }
  }
}`

// TestMockSalesforceConfigResolution verifies the mocksalesforce registry entry resolves with
// verification bypassed (unlike real Salesforce, whose non-bypassed path requires a
// ProviderApp row despite accepting every message).
func TestMockSalesforceConfigResolution(t *testing.T) { //nolint:paralleltest // mutates the provider catalog
	providers.SetupMockSalesforceProvider()

	info, err := providers.ReadInfo(providers.MockSalesforce)
	if err != nil {
		t.Fatalf("ReadInfo: %v", err)
	}

	cfg, err := GetProviderConfig("", ResolveProviderInfoAlias(info), deps.Dependencies{})
	if err != nil {
		t.Fatalf("GetProviderConfig: %v", err)
	}

	if !cfg.Verification.ShouldBypass() {
		t.Error("expected mocksalesforce webhook verification to be bypassed")
	}
}

// TestMockSalesforceEventParsing runs an EventBridge-wrapped CDC payload through the same
// object-payload dispatch the server's receive path uses and asserts Salesforce's real event
// implementation handles it: the envelope is unwrapped and the event fans out per record id.
func TestMockSalesforceEventParsing(t *testing.T) {
	t.Parallel()

	var rawEvent map[string]any
	if err := json.Unmarshal([]byte(mockSalesforceEventBridgePayload), &rawEvent); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	events, err := GetObjectTypeSubscribeEventsList(providers.MockSalesforce, rawEvent)
	if err != nil {
		t.Fatalf("GetObjectTypeSubscribeEventsList: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected recordIds to fan out into 2 events, got %d", len(events))
	}

	first := events[0]

	if evtType, err := first.EventType(); err != nil || evtType != common.SubscriptionEventTypeCreate {
		t.Errorf("expected create event, got %q (err %v)", evtType, err)
	}

	if objectName, err := first.ObjectName(); err != nil || objectName != "Account" {
		t.Errorf("expected object name %q, got %q (err %v)", "Account", objectName, err)
	}

	if rawName, err := first.RawEventName(); err != nil || rawName != "CREATE" {
		t.Errorf("expected raw event name %q, got %q (err %v)", "CREATE", rawName, err)
	}

	wantIDs := []string{"0015f00002J9YYEAA3", "0015f00002J9YYFAA3"}
	for i, event := range events {
		recordID, err := event.RecordId()
		if err != nil {
			t.Fatalf("RecordId event %d: %v", i, err)
		}

		if recordID != wantIDs[i] {
			t.Errorf("event %d: expected record id %q, got %q", i, wantIDs[i], recordID)
		}
	}
}
