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

var salesloftSubscriptionEvents = datautils.NewSet(
	"account_created",
	"account_updated",
	"account_deleted",

	"bulk_job_completed",

	"cadence_created",
	"cadence_updated",
	"cadence_deleted",

	"cadence_membership_created",
	"cadence_membership_updated",

	"call_created",
	"call_updated",

	"call_data_record_created",
	"call_data_record_updated",

	"conversation_created",

	// "conversation_recording_created", //we don't support read for this.

	// "conversation_transcript_created", //we don't support read for this obj.

	"email_updated",

	// "email_with_body_and_subject_updated", // Not supported object

	// "link_swap", //Not supported object

	"meeting_booked",
	"meeting_updated",

	"note_created",
	"note_updated",
	"note_deleted",

	// "person_created",
	// "person_updated", // Not supported object
	// "person_deleted",

	"step_created",
	"step_updated",
	"step_deleted",

	"success_created",

	"task_completed",
	"task_created",
	"task_updated",
	"task_deleted",

	"user_created",
	"user_updated",
)
