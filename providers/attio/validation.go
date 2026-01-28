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

func validateSubscriptionEvents(subscriptionEvents map[common.ObjectName]common.ObjectEvents) error {
	var validationErrors error

	for objectName, objectEvents := range subscriptionEvents {
		attioEvents, exist := attioObjectEvents[objectName]
		if !exist {
			validationErrors = errors.Join(validationErrors,
				fmt.Errorf("%s %w", objectName, errUnsupportedObject))

			continue
		}

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
	}

	return validationErrors
}
