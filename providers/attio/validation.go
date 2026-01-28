package attio

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

func validateRequest(params common.SubscribeParams) (*subscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*subscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T' got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: request is invalid: %w", errInvalidRequestType, err)
	}

	return req, nil
}

// validateSubscriptionEvents checks if the requested subscription events are valid.
// There are two types of objects that work differently (see Subscribe method for more details):
//
// Core Objects (lists, tasks, notes, workspace_members):
//   - Checks if the events are in our predefined list (attioObjectEvents)
//   - Makes sure the events actually exist for that object
//
// Standard/Custom Objects (people, companies, deals, etc.):
//   - Checks if the object exists in the workspace
//   - Only allows create, update, or delete events
func validateSubscriptionEvents(
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
	objectIDMap map[common.ObjectName]string,
) error {
	var validationErrors error

	for objectName, objectEvents := range subscriptionEvents {
		attioEvents, isCoreObject := attioObjectEvents[objectName]

		// PATTERN 1: Validate if its core objects
		if isCoreObject {
			// Get all supported events for this object
			supportedEvents := attioEvents.getAllSupportedEvents()

			supportedSet := make(map[providerEvent]bool)
			for _, evt := range supportedEvents {
				supportedSet[evt] = true
			}

			for _, event := range objectEvents.Events {
				providerEvents := attioEvents.toProviderEvents(event)

				if providerEvents == nil {
					validationErrors = errors.Join(validationErrors,
						fmt.Errorf("%w for object '%s'", errUnsupportedSubscriptionEvent, objectName))

					continue
				}

				if len(providerEvents) == 0 {
					validationErrors = errors.Join(validationErrors,
						fmt.Errorf("%w: event '%s' for object '%s'", errUnsupportedSubscriptionEvent, event, objectName))

					continue
				}

				// Validate that all provider events are supported
				for _, providerEvent := range providerEvents {
					if !supportedSet[providerEvent] {
						validationErrors = errors.Join(validationErrors,
							fmt.Errorf("%w: provider event '%s' for common event '%s' and object '%s'",
								errUnsupportedSubscriptionEvent, providerEvent, event, objectName))
					}
				}
			}
		} else {
			// PATTERN 2: Validate standard/custom objects
			_, exists := objectIDMap[objectName]
			if !exists {
				validationErrors = errors.Join(validationErrors,
					fmt.Errorf("object '%s' not supported or not activated in workspace", objectName))

				continue
			}

			for _, evt := range objectEvents.Events {
				switch evt {
				case common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete:

				// Valid
				default:
					validationErrors = errors.Join(validationErrors,
						fmt.Errorf("unsupported event '%s' for object '%s'", evt, objectName))
				}
			}
		}
	}

	return validationErrors
}
