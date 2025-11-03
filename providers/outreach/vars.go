package outreach

import "errors"

var (
	errInvalidRequestType   = errors.New("invalid request type")
	errMissingParams        = errors.New("missing required parameters")
	errUnsupportedEventType = errors.New("unsupported event type")
)
