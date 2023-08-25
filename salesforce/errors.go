package salesforce

import "errors"

var (
	ErrNotArray         = errors.New("records is not an array")
	ErrNotObject        = errors.New("record isn't an object")
	ErrNoFields         = errors.New("no fields specified")
	ErrNotString        = errors.New("nextRecordsUrl isn't a string")
	ErrNotBool          = errors.New("done isn't a boolean")
	ErrNotNumeric       = errors.New("totalSize isn't numeric")
	ErrMissingSubdomain = errors.New("missing Salesforce workspace name")
	ErrMissingClient    = errors.New("JSON http client not set")
)
