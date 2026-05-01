package subscriber

import (
	"context"
	"maps"
	"sort"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/microsoft/internal/webhook"
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

	// ReconciliationPlan describes the actions to take on each object to align RemoteSubscriptions
	// with EffectiveState (desired state).
	ReconciliationPlan struct {
		Create  map[common.ObjectName]common.ObjectEvents
		Refresh map[common.ObjectName]SubscriptionID
		Delete  []common.ObjectName
		// Extra is unique to the Microsoft Graph API. Single resource may have too many subscriptions.
		// Any excess should be cleaned up.
		Extra []SubscriptionID
	}
)

func newState(objectNames []ObjectName) State {
	events := make(State)
	for _, objectName := range objectNames {
		events[objectName] = common.ObjectEvents{}
	}

	return events
}

// Add copies the object events of the other to current.
func (s State) Add(other State) {
	maps.Copy(s, other)
}

// Equals reports if the subscription configuration matches between both states for the given object.
func (s State) Equals(other State, objectName ObjectName) bool {
	return webhook.NewChangeType(s[objectName].Events) == webhook.NewChangeType(other[objectName].Events)
}

// ReconcileTo partitions objects into three disjoint sets by comparing desired (effective)
// and actual (remote) state:
//
// - objectsToRemove: present remotely but not desired → DELETE all subscriptions for the object
// - objectsToCreate: desired but not present remotely → CREATE subscriptions for the object
// - objectsToUpdate: present in both → UPDATE existing subscriptions
//
// For Microsoft Graph, intersecting objects are always updated.
// Subscription.ExpirationDateTime must be continuously renewed,
// so there is no no-op path—every existing subscription requires an update.
//
// TODO The subscription may disappear under the hood as Microsoft auto cleans. Need to create a handler for this case.
func (r RemoteSubscriptions) ReconcileTo(state State) ReconciliationPlan {
	remoteObjects := datautils.NewSetFromList(RemoteSubsType(r).GetBuckets())
	desiredObjects := datautils.FromMap(state).KeySet()

	objectsToCreate := desiredObjects.Subtract(remoteObjects)
	objectsToRemove := remoteObjects.Subtract(desiredObjects)
	objectsToUpdate := remoteObjects.Intersection(desiredObjects)

	plan := ReconciliationPlan{
		Create:  make(map[common.ObjectName]common.ObjectEvents),
		Refresh: make(map[common.ObjectName]SubscriptionID),
		Delete:  objectsToRemove,
		Extra:   make([]SubscriptionID, 0),
	}

	for _, name := range objectsToCreate {
		plan.Create[name] = state[name]
	}

	remoteState := r.toState()
	for _, name := range objectsToUpdate {
		subscriptions := r[name]

		if len(subscriptions) == 0 {
			// Impossible. Remote state by definition must have at least one subscription for an object.
			continue
		}

		if len(subscriptions) > 1 {
			sort.Slice(subscriptions, func(i, j int) bool {
				return subscriptions[i].ExpirationDateTime.After(subscriptions[j].ExpirationDateTime)
			})
		}

		// Keep the first
		subscription := subscriptions[0]

		// Remove the rest
		for _, extra := range subscriptions[1:] {
			// There are multiple subscriptions associated with this object.
			// Keep only one of them, others must be removed.
			// This could happen due to user manually altering subscriptions
			// or any invalid state the connector has put the provider in.
			// This is highly unlikely but such possibility is left open.
			// Too many subscriptions for a single object. Remove the excess.
			plan.Extra = append(plan.Extra, extra.ID)
		}

		// Replace subscription with a more desired version.
		if remoteState.Equals(state, name) {
			plan.Refresh[name] = subscription.ID
		} else {
			// Create a new subscription which is different from the original.
			plan.Create[name] = state[name]
			// Mark old subscription for a cleanup.
			plan.Extra = append(plan.Extra, subscription.ID)
		}
	}

	return plan
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
