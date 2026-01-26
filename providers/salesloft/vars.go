package salesloft

import (
	"errors"

	"github.com/amp-labs/connectors/common"
)

var (
	errInvalidRequestType           = errors.New("invalid request type")
	errMissingParams                = errors.New("missing required parameters")
	errUnsupportedEventType         = errors.New("unsupported event type")
	errUnsupportedObject            = errors.New("unsupported object")
	errUnsupportedSubscriptionEvent = errors.New("unsupported subscription event")
	ErrMissingSignature             = errors.New("missing webhook signature header")
	ErrInvalidSignature             = errors.New("invalid webhook signature")
	//nolint:revive
)

// salesloftEventMappings maps the actual object names (used in read/write operations) to their webhook configurations.
// The map key (e.g., "tasks", "people", "activities/calls") is the object name supported by the connector.
// The ObjectName field in the value (e.g., "task", "person", "call") is the prefix used in webhook event names.
//
//nolint:gochecknoglobals
var salesloftEventMappings = map[common.ObjectName]salesloftObjectMapping{
	"accounts": {
		ObjectName: "account",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"account_created"},
			UpdateEvents: []moduleEvent{"account_updated"},
			DeleteEvents: []moduleEvent{"account_deleted"},
		},
	},

	"bulk_jobs": {
		ObjectName: "bulk_job",
		Events: eventMapping{
			UpdateEvents: []moduleEvent{"bulk_job_completed"},
		},
	},

	"cadences": {
		ObjectName: "cadence",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"cadence_created"},
			UpdateEvents: []moduleEvent{"cadence_updated"},
			DeleteEvents: []moduleEvent{"cadence_deleted"},
		},
	},

	"cadence_memberships": {
		ObjectName: "cadence_membership",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"cadence_membership_created"},
			UpdateEvents: []moduleEvent{"cadence_membership_updated"},
		},
	},

	"activities/calls": {
		ObjectName: "call",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"call_created"},
			UpdateEvents: []moduleEvent{"call_updated"},
		},
	},

	"call_data_records": {
		ObjectName: "call_data_record",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"call_data_record_created"},
			UpdateEvents: []moduleEvent{"call_data_record_updated"},
		},
	},

	"conversations": {
		ObjectName: "conversation",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"conversation_created"},
			UpdateEvents: []moduleEvent{"conversation_updated"},
		},
	},

	"activities/emails": {
		ObjectName: "email",
		Events: eventMapping{
			UpdateEvents: []moduleEvent{"email_updated"},
		},
	},

	"meetings": {
		ObjectName: "meeting",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"meeting_booked"},
			UpdateEvents: []moduleEvent{"meeting_updated"},
		},
	},

	"notes": {
		ObjectName: "note",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"note_created"},
			UpdateEvents: []moduleEvent{"note_updated"},
			DeleteEvents: []moduleEvent{"note_deleted"},
		},
	},

	"people": {
		ObjectName: "person",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"person_created"},
			UpdateEvents: []moduleEvent{"person_updated"},
			DeleteEvents: []moduleEvent{"person_deleted"},
		},
	},

	"steps": {
		ObjectName: "step",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"step_created"},
			UpdateEvents: []moduleEvent{"step_updated"},
			DeleteEvents: []moduleEvent{"step_deleted"},
		},
	},

	"successes": {
		ObjectName: "success",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"success_created"},
		},
	},

	"tasks": {
		ObjectName: "task",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"task_created"},
			UpdateEvents: []moduleEvent{"task_updated", "task_completed"},
			DeleteEvents: []moduleEvent{"task_deleted"},
		},
	},

	"users": {
		ObjectName: "user",
		Events: eventMapping{
			CreateEvents: []moduleEvent{"user_created"},
			UpdateEvents: []moduleEvent{"user_updated"},
		},
	},
}
