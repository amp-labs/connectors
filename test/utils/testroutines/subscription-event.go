package testroutines

import (
	"fmt"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

type SubscriptionEventExpected struct {
	Data SubscriptionEventExpectedData
	Err  SubscriptionEventExpectedErr
}

type SubscriptionEventExpectedData struct {
	EventType          common.SubscriptionEventType
	RawEventName       string
	ObjectName         string
	Workspace          string
	RecordId           string
	EventTimeStampNano int64
	UpdatedFields      []string
}

type SubscriptionEventExpectedErr struct {
	EventType          error
	RawEventName       error
	ObjectName         error
	Workspace          error
	RecordId           error
	EventTimeStampNano error
	UpdatedFields      error
}

type SubscriptionEventTestCase struct {
	Name                     string
	Input                    common.Event
	Expected                 []SubscriptionEventExpected
	SubscriptionEventListErr error
}

func (c SubscriptionEventTestCase) Run(t *testing.T) {
	t.Helper()

	result := testutils.NewCompareResult()

	switch event := c.Input.(type) {
	case common.CollapsedSubscriptionEvent:
		eventsList, err := event.SubscriptionEventList()
		if !result.AssertErr("SubscriptionEventList (err)", c.SubscriptionEventListErr, err) {
			break
		}

		if !result.Assert("Mismatching number of events", len(c.Expected), len(eventsList)) {
			break
		}

		for index, subscriptionEvent := range eventsList {
			result.Merge(validateSubscriptionEvent(t, subscriptionEvent, index, c.Expected[index]))
		}
	case common.SubscriptionEvent:
		if !result.Assert("expected multiple events, but got a single SubscriptionEvent", 1, len(c.Expected)) {
			break
		}
		result.Merge(validateSubscriptionEvent(t, event, 0, c.Expected[0]))
	case common.SubscriptionUpdateEvent:
		if !result.Assert("expected multiple events, but got a single SubscriptionUpdateEvent", 1, len(c.Expected)) {
			break
		}
		result.Merge(validateSubscriptionUpdateEvent(event, c.Expected[0]))
	default:
		result.AddDiff("Input is of unknown type %T", event)
	}

	result.Validate(t, c.Name)
}

func validateSubscriptionUpdateEvent(
	event common.SubscriptionUpdateEvent,
	expected SubscriptionEventExpected,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	fields, err := event.UpdatedFields()
	result.AssertErr("UpdatedFields (err)", expected.Err.UpdatedFields, err)
	result.Assert("UpdatedFields", expected.Data.UpdatedFields, fields)

	return result
}

func validateSubscriptionEvent(
	t *testing.T, event common.SubscriptionEvent, arrPosition int, expected SubscriptionEventExpected,
) *testutils.CompareResult {
	t.Helper()

	result := testutils.NewCompareResult()

	// Test EventType
	eventType, err := event.EventType()
	result.AssertErr(fmt.Sprintf("[%v].EventType (err)", arrPosition), expected.Err.EventType, err)
	result.Assert(fmt.Sprintf("[%v].EventType", arrPosition), expected.Data.EventType, eventType)

	// Test RawEventName
	rawEventName, err := event.RawEventName()
	result.AssertErr(fmt.Sprintf("[%v].RawEventName (err)", arrPosition), expected.Err.RawEventName, err)
	result.Assert(fmt.Sprintf("[%v].RawEventName", arrPosition), expected.Data.RawEventName, rawEventName)

	// Test ObjectName
	objectName, err := event.ObjectName()
	result.AssertErr(fmt.Sprintf("[%v].ObjectName (err)", arrPosition), expected.Err.ObjectName, err)
	result.Assert(fmt.Sprintf("[%v].ObjectName", arrPosition), expected.Data.ObjectName, objectName)

	// Test Workspace
	workspace, err := event.Workspace()
	result.AssertErr(fmt.Sprintf("[%v].Workspace (err)", arrPosition), expected.Err.Workspace, err)
	result.Assert(fmt.Sprintf("[%v].Workspace", arrPosition), expected.Data.Workspace, workspace)

	// Test RecordId
	recordID, err := event.RecordId()
	result.AssertErr(fmt.Sprintf("[%v].RecordId (err)", arrPosition), expected.Err.RecordId, err)
	result.Assert(fmt.Sprintf("[%v].RecordId", arrPosition), expected.Data.RecordId, recordID)

	// Test EventTimeStampNano
	timestamp, err := event.EventTimeStampNano()
	result.AssertErr(
		fmt.Sprintf("[%v].EventTimeStampNano (err)", arrPosition), expected.Err.EventTimeStampNano, err)
	result.Assert(
		fmt.Sprintf("[%v].EventTimeStampNano", arrPosition), expected.Data.EventTimeStampNano, timestamp)

	return result
}
