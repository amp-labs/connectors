package mockutils

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

var SubscriptionResultComparator = subscriptionResultComparator{}

type subscriptionResultComparator struct{}

func (subscriptionResultComparator) CompareWithoutResultArg(
	actual, expected *common.SubscriptionResult,
) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	result.Assert("ObjectEvents", expected.ObjectEvents, actual.ObjectEvents)
	result.Assert("Status", expected.Status, actual.Status)
	result.Assert("Objects", expected.Objects, actual.Objects)
	result.Assert("Events", expected.Events, actual.Events)
	result.Assert("UpdateFields", expected.UpdateFields, actual.UpdateFields)
	result.Assert("PassThroughEvents", expected.PassThroughEvents, actual.PassThroughEvents)

	return result
}
