package subscriber

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

type (
	// RemoteSubscriptions describes the actual state of subscriptions in the remote system.
	// Each ObjectName maps to a list of remote SubscriptionResource items.
	RemoteSubscriptions RemoteSubsType
	RemoteSubsType      = datautils.IndexedLists[ObjectName, SubscriptionResource]

	// State describes the set of object events that the connector should
	// try to achieve (desired) and that are also reported back as the observed outcome.
	//
	// It is part of the connector's input/output contract: the caller expresses intent
	// with State, and the connector returns the same type reflecting what actually
	// happened (which may differ on failure or rollback).
	State map[ObjectName]common.ObjectEvents
)

func newState(objectNames []ObjectName) State {
	events := make(State)
	for _, objectName := range objectNames {
		events[objectName] = common.ObjectEvents{}
	}

	return events
}

func (r RemoteSubscriptions) toState() State {
	result := make(State)

	for objectName, subscriptions := range r {
		eventSet := datautils.NewSet[common.SubscriptionEventType]()

		for _, subscription := range subscriptions {
			eventSet.Add(subscription.ChangeType.EventTypes())
		}

		// List of events may be nil.
		var events []common.SubscriptionEventType
		if list := eventSet.List(); len(list) != 0 {
			events = list
		}

		result[objectName] = common.ObjectEvents{
			Events:            events,
			WatchFields:       nil,
			WatchFieldsAll:    false,
			PassThroughEvents: nil,
		}
	}

	return result
}

// fetchSubscriptions retrieves the current Microsoft Graph change‑notification subscriptions for the given objects.
//
// nolint:lll
// According to https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#subscription-request,
// Graph should guard against duplicate subscriptions (same changeType and resource), but in practice
// multiple subscriptions for the same combination can exist. Therefore, before blindly creating,
// updating, or deleting subscriptions, this method first reads the current state and allows the caller
// to reconcile the desired vs. actual subscriptions.
func (s Strategy) fetchSubscriptions(ctx context.Context) (RemoteSubscriptions, error) {
	url, err := s.getSubscriptionURL()
	if err != nil {
		return nil, err
	}

	response, err := s.client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	subscriptions, err := common.UnmarshalJSON[subscriptionResources](response)
	if err != nil {
		return nil, err
	}

	result := make(RemoteSubscriptions)

	for _, resource := range subscriptions.List {
		objectName := ObjectName(resource.Resource)
		RemoteSubsType(result).Add(objectName, resource)
	}

	return result, nil
}

// subscriptionResources is the output of "GET /subscriptions".
// https://learn.microsoft.com/en-us/graph/api/subscription-list?view=graph-rest-1.0&tabs=http#response-1
type subscriptionResources struct {
	List []SubscriptionResource `json:"value"`
}
