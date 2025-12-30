package stripe

import "errors"

var (
	errInvalidRequestType     = errors.New("invalid request type")
	errMissingParams          = errors.New("missing required parameters")
	errUnsupportedEventType   = errors.New("unsupported event type")
	errObjectTypeMismatch     = errors.New("object type mismatch")
	errObjectFieldNotFound    = errors.New("object field not found in metadata")
	errNoValuesDefined        = errors.New("no values defined for object field in metadata")
	errNotYetImplemented      = errors.New("not yet implemented")
	errMissingSignature       = errors.New("missing webhook signature header")
	errInvalidSignature       = errors.New("invalid webhook signature")
	errInvalidEventTypeFormat = errors.New("invalid event type format")
	errMissingTimestamp       = errors.New("missing timestamp")
	errNoSignaturesFound      = errors.New("no signatures found")
)
