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

//nolint:gochecknoglobals
var salesloftEventMappings = map[common.ObjectName]SalesloftObjectMapping{
	"accounts": {
		ObjectName: "account",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"account_created"},
			UpdateEvents: []ModuleEvent{"account_updated"},
			DeleteEvents: []ModuleEvent{"account_deleted"},
		},
	},

	"bulk_jobs": {
		ObjectName: "bulk_job",
		Events: EventMapping{
			UpdateEvents: []ModuleEvent{"bulk_job_completed"},
		},
	},

	"cadences": {
		ObjectName: "cadence",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"cadence_created"},
			UpdateEvents: []ModuleEvent{"cadence_updated"},
			DeleteEvents: []ModuleEvent{"cadence_deleted"},
		},
	},

	"cadence_memberships": {
		ObjectName: "cadence_membership",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"cadence_membership_created"},
			UpdateEvents: []ModuleEvent{"cadence_membership_updated"},
		},
	},

	"activities/calls": {
		ObjectName: "call",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"call_created"},
			UpdateEvents: []ModuleEvent{"call_updated"},
		},
	},

	"call_data_records": {
		ObjectName: "call_data_record",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"call_data_record_created"},
			UpdateEvents: []ModuleEvent{"call_data_record_updated"},
		},
	},

	"conversations": {
		ObjectName: "conversation",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"conversation_created"},
			UpdateEvents: []ModuleEvent{"conversation_updated"},
		},
	},

	"activities/emails": {
		ObjectName: "email",
		Events: EventMapping{
			UpdateEvents: []ModuleEvent{"email_updated"},
		},
	},

	"meetings": {
		ObjectName: "meeting",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"meeting_booked"},
			UpdateEvents: []ModuleEvent{"meeting_updated"},
		},
	},

	"notes": {
		ObjectName: "note",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"note_created"},
			UpdateEvents: []ModuleEvent{"note_updated"},
			DeleteEvents: []ModuleEvent{"note_deleted"},
		},
	},

	"people": {
		ObjectName: "person",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"person_created"},
			UpdateEvents: []ModuleEvent{"person_updated"},
			DeleteEvents: []ModuleEvent{"person_deleted"},
		},
	},

	"steps": {
		ObjectName: "step",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"step_created"},
			UpdateEvents: []ModuleEvent{"step_updated"},
			DeleteEvents: []ModuleEvent{"step_deleted"},
		},
	},

	"successes": {
		ObjectName: "success",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"success_created"},
		},
	},

	"tasks": {
		ObjectName: "task",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"task_created"},
			UpdateEvents: []ModuleEvent{"task_updated", "task_completed"},
			DeleteEvents: []ModuleEvent{"task_deleted"},
		},
	},

	"users": {
		ObjectName: "user",
		Events: EventMapping{
			CreateEvents: []ModuleEvent{"user_created"},
			UpdateEvents: []ModuleEvent{"user_updated"},
		},
	},
}
