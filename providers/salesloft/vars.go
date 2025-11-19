package salesloft

import (
	"errors"

	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	errInvalidRequestType    = errors.New("invalid request type")
	errMissingParams         = errors.New("missing required parameters")
	errWatchFieldsAll        = errors.New("watch fields all is not supported")
	errTooManyWatchFields    = errors.New("too many watch fields")
	errSubscriptionFailed    = errors.New("subscription failed")
	errNoSubscriptionCreated = errors.New("no subscription created")
	errUnsupportedEventType  = errors.New("unsupported event type")
	errFieldNotFound         = errors.New("field not found")
	errObjectNameNotFound    = errors.New("object name not found")
	errInvalidModuleEvent    = errors.New("invalid module event")
	//nolint:revive
	errInconsistentChannelIdsMismatch = errors.New("all events must have the same channel id")
	errChannelIdMismatch              = errors.New("channel id does not match provided unique ref")
	errInvalidDuration                = errors.New("duration cannot be greater than 1 week")
	errModuleNameNotString            = errors.New("module_name is not a string")
	errAPINameNotString               = errors.New("api_name is not a string")
	errIDNotString                    = errors.New("id is not a string")
	errFieldIDNotString               = errors.New("field id is not a string")
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

	"conversation_recording_created",

	"conversation_transcript_created",

	"email_updated",

	"email_with_body_and_subject_updated",

	"link_swap",

	"meeting_booked",
	"meeting_updated",

	"note_created",
	"note_updated",
	"note_deleted",

	"person_created",
	"person_updated",
	"person_deleted",

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
