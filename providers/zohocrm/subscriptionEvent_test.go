package zohocrm

import (
	"encoding/json"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/tools/debug"
	"gotest.tools/v3/assert"
)

//nolint:funlen,cyclop
func TestCollapsedSubscriptionEvent(t *testing.T) {
	t.Parallel()
	logger := t.Log

	var evt *CollapsedSubscriptionEvent

	evtStr := `{
			"server_time": 1750102639787,
			"affected_values": [
				{
					"record_id": "6756839000000575405",
					"values": {
						"Company": "Rangoni Of Test",
						"Phone": "555-555-1111"
					}
				}
			],
			"query_params": {},
			"module": "Leads",
			"resource_uri": "https://www.zohoapis.com/crm/v2/Leads",
			"ids": [
				"6756839000000575405"
			],
			"affected_fields": [
				{
					"6756839000000575405": [
						"Company",
						"Phone"
					]
				}
			],
			"operation": "update",
			"channel_id": "1105420521999070702",
			"token": "c3504777-db15-4332-8286-478a1b5006bc"
		}
`

	err := json.Unmarshal([]byte(evtStr), &evt)
	if err != nil {
		t.Fatalf("failed to unmarshal evt: %v", err)
	}

	// logger("evt", debug.PrettyFormatStringJSON(evt))

	subevts, err := evt.SubscriptionEventList()
	if err != nil {
		t.Fatalf("failed to get subscription event list: %v", err)
	}

	logger("evts", debug.PrettyFormatStringJSON(subevts))

	subevt := subevts[0]

	evtType, err := subevt.EventType()
	if err != nil {
		t.Fatalf("failed to get updated fields: %v", err)
	}

	assert.Equal(t, evtType, common.SubscriptionEventTypeUpdate)

	logger("evtType", evtType)

	rawEventName, err := subevt.RawEventName()
	if err != nil {
		t.Fatalf("failed to get raw event name: %v", err)
	}

	assert.Equal(t, rawEventName, "update")

	logger("rawEventName", rawEventName)

	objectName, err := subevt.ObjectName()
	if err != nil {
		t.Fatalf("failed to get object name: %v", err)
	}

	assert.Equal(t, objectName, "Leads")

	logger("objectName", objectName)

	workspace, err := subevt.Workspace()
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}

	assert.Equal(t, workspace, "")

	logger("workspace", workspace)

	recordId, err := subevt.RecordId()
	if err != nil {
		t.Fatalf("failed to get record id: %v", err)
	}

	logger("recordId", recordId)

	evtTimeStampNano, err := subevt.EventTimeStampNano()
	if err != nil {
		t.Fatalf("failed to get event time stamp nano: %v", err)
	}

	logger("evtTimeStampNano", evtTimeStampNano)

	subevtUpdateEvent, ok := subevt.(common.SubscriptionUpdateEvent)
	if !ok {
		t.Fatalf("failed to cast to subscription update event")
	}

	updatedFields, err := subevtUpdateEvent.UpdatedFields()
	if err != nil {
		t.Fatalf("failed to get updated fields: %v", err)
	}

	logger("updatedFields", updatedFields)
}
