// Package subscriptionhelper provides utilities for categorizing object subscription events
// based on their existing and desired states. This allows determining the appropriate
// action (create, keep, update, remove) for each object subscription.
package subscriptionhelper

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

type (
	ObjectName   = common.ObjectName
	ObjectEvents = common.ObjectEvents
)

// EventSegments categorizes subscription events into four action categories:
//   - ToCreate: Objects that exist in desired state but not in existing state
//   - ToKeep: Objects that exist in both states with identical event configurations
//   - ToUpdate: Objects that exist in both states but with different event configurations
//   - ToRemove: Objects that exist in existing state but not in desired state
type EventSegments struct {
	ToCreate datautils.Map[ObjectName, ObjectEvents]
	ToKeep   datautils.Map[ObjectName, ObjectEvents]
	ToUpdate datautils.Map[ObjectName, ObjectEvents]
	ToRemove datautils.Map[ObjectName, ObjectEvents]
}

// SegmentSubscriptionEvents compares previous (existing) and new (desired) object subscription events
// and categorizes them into EventSegments based on the required action for each object.
//
// The categorization logic:
//   - Objects in newEvents but not in prevEvents → ToCreate
//   - Objects in prevEvents but not in newEvents → ToRemove
//   - Objects in both with equal event configuration → ToKeep
//   - Objects in both with different event configuration → ToUpdate
//
// Parameters:
//   - prevEvents: The existing subscription events (current state)
//   - newEvents: The desired subscription events (target state)
//
// Returns:
//   - EventSegments containing the categorized events for each action type
func SegmentSubscriptionEvents(
	prevEvents datautils.Map[ObjectName, ObjectEvents],
	newEvents datautils.Map[ObjectName, ObjectEvents],
) EventSegments {
	prevEventsMap := datautils.FromMap(prevEvents)
	newEventsMap := datautils.FromMap(newEvents)

	currentObjects := prevEventsMap.KeySet()
	desiredObjects := newEventsMap.KeySet()

	objectsToCreate := desiredObjects.Subtract(currentObjects)
	objectsToRemove := currentObjects.Subtract(desiredObjects)
	objectsIntersection := currentObjects.Intersection(desiredObjects)

	toCreate := newEventsMap.ShallowSubset(objectsToCreate)
	toRemove := prevEventsMap.ShallowSubset(objectsToRemove)
	toKeep := make(datautils.Map[ObjectName, ObjectEvents])
	toUpdate := make(datautils.Map[ObjectName, ObjectEvents])

	// For object intersection we need to sort them into Keep or Update categories.
	// If values are the same then we mark them as to Keep, otherwise some Update is needed.
	for _, name := range objectsIntersection {
		if prevEvents[name].Equals(newEvents[name]) {
			toKeep[name] = newEvents[name]
		} else {
			toUpdate[name] = newEvents[name]
		}
	}

	return EventSegments{
		ToCreate: toCreate,
		ToKeep:   toKeep,
		ToUpdate: toUpdate,
		ToRemove: toRemove,
	}
}
