package mocksub

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// ObjectIDIndexResolver returns an ObjectNameFromEventFunc mirroring Attio's
// GetObjectNameFromEvent: standard/custom-object changes arrive as generic record.* events that
// identify their object only by an id.object_id UUID, which the real connector maps to an
// object name by fetching the workspace's objects. The mock resolves the same id.object_id
// against the store's seeded object-name index instead (SeedObjectName). Events whose id map
// carries no object_id (core-object events like note.*, task.*) fall back to their own
// ObjectName, exactly as the real resolver does.
func ObjectIDIndexResolver(store *Store) ObjectNameFromEventFunc {
	return func(_ context.Context, event common.SubscriptionEvent) (string, error) {
		raw, err := event.RawMap()
		if err != nil {
			return "", fmt.Errorf("getting raw event map: %w", err)
		}

		idMap, ok := raw["id"].(map[string]any)
		if !ok {
			return "", fmt.Errorf("%w: expected id to be map[string]any, got %T", common.ErrBadRequest, raw["id"])
		}

		objectID, ok := idMap["object_id"].(string)
		if !ok {
			// No object_id: a core-object event whose object is in the event type.
			return event.ObjectName()
		}

		name, found := store.ObjectNameFor(objectID)
		if !found {
			return "", fmt.Errorf("%w: object id %q not seeded in object-name index", common.ErrNotFound, objectID)
		}

		return name, nil
	}
}
