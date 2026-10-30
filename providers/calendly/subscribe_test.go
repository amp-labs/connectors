package calendly

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/require"
)

func TestBuildCalendlyEventList_eventTypes(t *testing.T) {
	t.Parallel()

	req := &SubscriptionRequest{}

	params := common.SubscribeParams{
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"event_types": {
				Events: []common.SubscriptionEventType{
					common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete,
				},
			},
		},
	}

	got := buildCalendlyEventList(params, req)
	require.ElementsMatch(t, []string{
		"event_type.created",
		"event_type.updated",
		"event_type.deleted",
	}, got)
}

func TestSplitEventName(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"event_type", "created"}, splitEventName("event_type.created"))
	require.Equal(t, []string{"routing_form_submission", "created"}, splitEventName("routing_form_submission.created"))
}
