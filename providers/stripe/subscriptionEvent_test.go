package stripe

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSubscriptionEvents(t *testing.T) {
	t.Parallel()

	data := testutils.DataFromFile(t, "subscribe/setup_intent-created.json")

	var evt CollapsedSubscriptionEvent

	err := json.Unmarshal(data, &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	tests := []testconn.TestCaseSubscriptionEvent{
		{
			Name:  "Setup intent created event",
			Input: evt,
			Expected: []testconn.SubscriptionEventExpected{{
				Data: testconn.SubscriptionEventExpectedData{
					EventType:          common.SubscriptionEventTypeCreate,
					RawEventName:       "setup_intent.created",
					ObjectName:         "setup_intent",
					Workspace:          "",
					RecordId:           "seti_1NG8Du2eZvKYlo2C9XMqbR0x",
					EventTimeStampNano: time.Unix(1686089970, 0).UnixNano(),
					UpdatedFields:      []string{},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t)
		})
	}
}
