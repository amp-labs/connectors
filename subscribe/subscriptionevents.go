package subscribe

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/mitchellh/mapstructure"
)

// errSubscriptionEventCasterNotDeclared is returned by VerificationConfig.CastEvents when the
// provider's VerificationConfig declares no eventCaster.
var errSubscriptionEventCasterNotDeclared = errors.New("subscription event caster not declared for provider")

// SubscriptionEventCaster converts a list of generic maps into a list of typed
// SubscriptionEvent implementations for a specific provider. Declared per-provider via
// VerificationConfig.eventCaster (see the per-provider configs).
type SubscriptionEventCaster func(list []map[string]any) ([]common.SubscriptionEvent, error)

// CastSubscriptionEvents wraps CastArray with the conversion to []common.SubscriptionEvent.
func CastSubscriptionEvents[T common.SubscriptionEvent](
	list []map[string]any,
) ([]common.SubscriptionEvent, error) {
	typedList, err := CastArray[T](list)
	if err != nil {
		return nil, fmt.Errorf("failed to cast to array: %w", err)
	}

	result := make([]common.SubscriptionEvent, len(typedList))
	for i, evt := range typedList {
		result[i] = evt
	}

	return result, nil
}

// CastArray converts a slice of generic maps to a slice of typed structs using mapstructure.
// Generic function that works for any struct type T.
func CastArray[T any](list []map[string]any) ([]T, error) {
	result := make([]T, len(list))

	for i, item := range list {
		err := mapstructure.Decode(item, &result[i])
		if err != nil {
			return nil, fmt.Errorf("failed to cast to %T: %w", new(T), err)
		}
	}

	return result, nil
}
