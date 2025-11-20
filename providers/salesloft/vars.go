package salesloft

import (
	"errors"

	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	errInvalidRequestType   = errors.New("invalid request type")
	errMissingParams        = errors.New("missing required parameters")
	errUnsupportedEventType = errors.New("unsupported event type")

	//nolint:revive
)

var salesloftEventMappings = map[string]SalesloftEventMapping{
	"accounts": {
		ObjectName: "account",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("account_created"),
			SalesloftEventType("account_updated"),
			SalesloftEventType("account_deleted"),
		),
	},

	"bulk_jobs": {
		ObjectName: "bulk_job",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("bulk_job_completed"),
		),
	},

	"cadences": {
		ObjectName: "cadence",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("cadence_created"),
			SalesloftEventType("cadence_updated"),
			SalesloftEventType("cadence_deleted"),
		),
	},

	"cadence_memberships": {
		ObjectName: "cadence_membership",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("cadence_membership_created"),
			SalesloftEventType("cadence_membership_updated"),
		),
	},

	"activities/calls": {
		ObjectName: "call",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("call_created"),
			SalesloftEventType("call_updated"),
		),
	},

	"call_data_records": {
		ObjectName: "call_data_record",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("call_data_record_created"),
			SalesloftEventType("call_data_record_updated"),
		),
	},

	"conversations": {
		ObjectName: "conversation",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("conversation_created"),
			SalesloftEventType("conversation_updated"),
		),
	},

	"activities/emails": {
		ObjectName: "email",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("email_updated"),
		),
	},

	"meetings": {
		ObjectName: "meeting",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("meeting_booked"),
			SalesloftEventType("meeting_updated"),
		),
	},

	"notes": {
		ObjectName: "note",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("note_created"),
			SalesloftEventType("note_updated"),
			SalesloftEventType("note_deleted"),
		),
	},

	"people": {
		ObjectName: "person",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("person_created"),
			SalesloftEventType("person_updated"),
			SalesloftEventType("person_deleted"),
		),
	},

	"steps": {
		ObjectName: "step",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("step_created"),
			SalesloftEventType("step_updated"),
			SalesloftEventType("step_deleted"),
		),
	},

	"successes": {
		ObjectName: "success",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("success_created"),
		),
	},

	"tasks": {
		ObjectName: "task",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("task_created"),
			SalesloftEventType("task_updated"),
			SalesloftEventType("task_deleted"),
			SalesloftEventType("task_completed"),
		),
	},
	"users": {
		ObjectName: "user",
		SupportedEvents: datautils.NewSet(
			SalesloftEventType("user_created"),
			SalesloftEventType("user_updated"),
		),
	},
}
