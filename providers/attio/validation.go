package attio

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
)

func validateRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
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
//   - Checks if the object exists in the standardObjects fetched via Attio API
//   - Only allows create, update, or delete events

// nolint:funlen, cyclop, gocognit
func validateSubscriptionEvents(
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
	standardObjects map[common.ObjectName]string,
) error {
	if len(subscriptionEvents) == 0 {
		return fmt.Errorf("%w: subscription events are empty", errMissingParams)
	}

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

				if len(providerEvents) == 0 {
					validationErrors = errors.Join(validationErrors,
						fmt.Errorf("%w for object '%s'", errUnsupportedSubscriptionEvent, objectName))

					continue
				}
				// Validate that all provider events are supported
				for _, providerEvent := range providerEvents {
					if !supportedSet[providerEvent] {
						validationErrors = errors.Join(validationErrors,
							fmt.Errorf("%w: provider event '%s' for common event '%s' and object '%s'",
								errUnsupportedSubscriptionEvent, providerEvent, event, objectName))

						continue
					}
				}
			}
		} else {
			// PATTERN 2: Validate standard/custom objects
			_, exists := standardObjects[objectName]
			if !exists {
				validationErrors = errors.Join(validationErrors,
					fmt.Errorf("%s: %w", objectName, errObjectNotFound))

				continue
			}

			for _, evt := range objectEvents.Events {
				// We only support create, update, delete events for standard/custom objects
				//nolint:exhaustive
				switch evt {
				case common.SubscriptionEventTypeCreate,
					common.SubscriptionEventTypeUpdate,
					common.SubscriptionEventTypeDelete:

				// Valid
				default:
					validationErrors = errors.Join(validationErrors,
						fmt.Errorf("%w: event '%s' for object '%s'", errUnsupportedSubscriptionEvent, evt, objectName))

					continue
				}
			}
		}
	}

	return validationErrors
}
