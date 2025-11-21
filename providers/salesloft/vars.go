package salesloft

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	errInvalidRequestType   = errors.New("invalid request type")
	errMissingParams        = errors.New("missing required parameters")
	errUnsupportedEventType = errors.New("unsupported event type")
	errUnsupportedObject    = errors.New("unsupported object")
	//nolint:revive
)

//nolint:gochecknoglobals
var salesloftEventMappings = map[common.ObjectName]SalesloftEventMapping{
	"accounts": {
		ObjectName: "account",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("account_created"),
			ModuleEvent("account_updated"),
			ModuleEvent("account_deleted"),
		),
	},

	"bulk_jobs": {
		ObjectName: "bulk_job",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("bulk_job_completed"),
		),
	},

	"cadences": {
		ObjectName: "cadence",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("cadence_created"),
			ModuleEvent("cadence_updated"),
			ModuleEvent("cadence_deleted"),
		),
	},

	"cadence_memberships": {
		ObjectName: "cadence_membership",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("cadence_membership_created"),
			ModuleEvent("cadence_membership_updated"),
		),
	},

	"activities/calls": {
		ObjectName: "call",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("call_created"),
			ModuleEvent("call_updated"),
		),
	},

	"call_data_records": {
		ObjectName: "call_data_record",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("call_data_record_created"),
			ModuleEvent("call_data_record_updated"),
		),
	},

	"conversations": {
		ObjectName: "conversation",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("conversation_created"),
			ModuleEvent("conversation_updated"),
		),
	},

	"activities/emails": {
		ObjectName: "email",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("email_updated"),
		),
	},

	"meetings": {
		ObjectName: "meeting",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("meeting_booked"),
			ModuleEvent("meeting_updated"),
		),
	},

	"notes": {
		ObjectName: "note",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("note_created"),
			ModuleEvent("note_updated"),
			ModuleEvent("note_deleted"),
		),
	},

	"people": {
		ObjectName: "person",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("person_created"),
			ModuleEvent("person_updated"),
			ModuleEvent("person_deleted"),
		),
	},

	"steps": {
		ObjectName: "step",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("step_created"),
			ModuleEvent("step_updated"),
			ModuleEvent("step_deleted"),
		),
	},

	"successes": {
		ObjectName: "success",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("success_created"),
		),
	},

	"tasks": {
		ObjectName: "task",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("task_created"),
			ModuleEvent("task_updated"),
			ModuleEvent("task_deleted"),
			ModuleEvent("task_completed"),
		),
	},
	"users": {
		ObjectName: "user",
		SupportedEvents: datautils.NewSet(
			ModuleEvent("user_created"),
			ModuleEvent("user_updated"),
		),
	},
}
