package attio

import (
	"errors"

	"github.com/amp-labs/connectors/common"
)

var (
	errInvalidRequestType           = errors.New("invalid request type")
	errMissingParams                = errors.New("missing required parameters")
	errUnsupportedObject            = errors.New("unsupported object")
	ErrMissingSignature             = errors.New("missing webhook signature header")
	ErrInvalidSignature             = errors.New("invalid webhook signature")
	errUnsupportedSubscriptionEvent = errors.New("unsupported subscription event")
)

//nolint:gochecknoglobals
var attioObjectEvents = map[common.ObjectName]objectEvents{
	"lists": {
		createEvents: []providerEvent{"list.created"},
		updateEvents: []providerEvent{"list.updated"},
		deleteEvents: []providerEvent{"list.deleted"},
	},

	"workspace_members": {
		createEvents: []providerEvent{"workspace-member.created"},
		updateEvents: []providerEvent{},
		deleteEvents: []providerEvent{},
	},

	"tasks": {
		createEvents: []providerEvent{"task.created"},
		updateEvents: []providerEvent{"task.updated"},
		deleteEvents: []providerEvent{"task.deleted"},
	},
	"notes": {
		createEvents: []providerEvent{"note.created"},
		updateEvents: []providerEvent{"note.updated", "note-content.updated"},
		deleteEvents: []providerEvent{"note.deleted"},
	},
}
