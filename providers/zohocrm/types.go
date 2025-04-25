package zohocrm

import "errors"

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

var (
	errInvalidRequestType = errors.New("invalid request type")
	errMissingParams      = errors.New("missing required parameters")
	errInvalidField       = errors.New("invalid field format")
	errValuesIdMismatch   = errors.New("record id and affected values record id does not match")
	errInvalidResponse    = errors.New("invalid response format")
)

// WatchResponse represents the top-level response from the Zoho CRM watch API.
type WatchResponse struct {
	Watch []WatchResult `json:"watch"`
}

// WatchResult represents a single watch subscription result.
type WatchResult struct {
	Code    string       `json:"code"`
	Details WatchDetails `json:"details"`
	Message string       `json:"message"`
	Status  string       `json:"status"`
}

// WatchDetails contains the details of the watch subscription.
type WatchDetails struct {
	Events []WatchEvent `json:"events"`
}

// WatchEvent represents a single event in the watch subscription.
//
//nolint:tagliatelle
type WatchEvent struct {
	ChannelExpiry string `json:"channel_expiry"`
	ResourceURI   string `json:"resource_uri"`
	ResourceID    string `json:"resource_id"`
	ResourceName  string `json:"resource_name"`
	ChannelID     string `json:"channel_id"`
}
