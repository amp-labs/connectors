package zohocrm

import (
	"errors"
	"time"
)

const (
	OperationCreate = "create"
	OperationEdit   = "edit"
	OperationDelete = "delete"
	OperationAll    = "all"

	maxWatchFields = 10

	ResultStatusSuccess = "SUCCESS"
	defaultDuration     = 7 * 24 * time.Hour // 1 week and this is max duration for subscription
)

var (
	errInvalidRequestType    = errors.New("invalid request type")
	errMissingParams         = errors.New("missing required parameters")
	errInvalidField          = errors.New("invalid field format")
	errValuesIdMismatch      = errors.New("record id and affected values record id does not match")
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
)

// uniqueFields maps the fields to the uniquely required fields.
var uniqueFields = map[string]string{ // nolint:gochecknoglobals
	"sic_code":                 "SIC_Code",
	"skype_id":                 "Skype_ID",
	"num_sent":                 "Num_sent",
	"what_id":                  "What_Id",
	"who_id":                   "Who_Id",
	"all_day":                  "All_day",
	"zip_code":                 "ZIP_Code",
	"cti_entry":                "CTI_Entry",
	"call_duration_in_seconds": "Call_Duration_in_seconds",
	"caller_id":                "Caller_ID",
	"scheduled_in_crm":         "Scheduled_In_CRM",
}
