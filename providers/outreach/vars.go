package outreach

import "errors"

var (
	errInvalidRequestType   = errors.New("invalid request type")
	errMissingParams        = errors.New("missing required parameters")
	errUnsupportedEventType = errors.New("unsupported event type")

	ErrMissingSignature                = errors.New("missing webhook signature header")
	ErrInvalidSignature                = errors.New("invalid webhook signature")
	errUnexpectedSubscriptionEventType = errors.New("unexpected subscription event type")
)
