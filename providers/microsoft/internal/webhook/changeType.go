package webhook

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

// ChangeType can be created/updated/deleted or a combination of them.
type ChangeType string

const (
	ChangeTypeCreated = "created"
	ChangeTypeUpdated = "updated"
	ChangeTypeDeleted = "deleted"
)

func NewChangeType(eventTypes []common.SubscriptionEventType) ChangeType {
	result := make([]string, 0, 3) // nolint:mnd
	requestedEvents := datautils.NewSetFromList(eventTypes)

	for _, item := range []datautils.Pair[common.SubscriptionEventType, string]{
		{Left: common.SubscriptionEventTypeCreate, Right: ChangeTypeCreated},
		{Left: common.SubscriptionEventTypeUpdate, Right: ChangeTypeUpdated},
		{Left: common.SubscriptionEventTypeDelete, Right: ChangeTypeDeleted},
	} {
		if requestedEvents.Has(item.Left) {
			result = append(result, item.Right)
		}
	}

	return ChangeType(strings.Join(result, ","))
}

func (c ChangeType) EventTypes() []common.SubscriptionEventType {
	parts := strings.Split(string(c), ",")
	result := make([]common.SubscriptionEventType, len(parts))

	for index, part := range parts {
		switch part {
		case ChangeTypeCreated:
			result[index] = common.SubscriptionEventTypeCreate
		case ChangeTypeUpdated:
			result[index] = common.SubscriptionEventTypeUpdate
		case ChangeTypeDeleted:
			result[index] = common.SubscriptionEventTypeDelete
		default:
			result[index] = common.SubscriptionEventTypeOther
		}
	}

	return result
}
