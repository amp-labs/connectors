package stripe

import "errors"

var (
	errInvalidRequestType      = errors.New("invalid request type")
	errMissingParams           = errors.New("missing required parameters")
	errUnsupportedEventType    = errors.New("unsupported event type")
	errObjectTypeMismatch      = errors.New("object type mismatch")
	errObjectFieldNotFound     = errors.New("object field not found in metadata")
	errNoValuesDefined         = errors.New("no values defined for object field in metadata")
	errMissingSignature        = errors.New("missing webhook signature header")
	errInvalidSignature        = errors.New("invalid webhook signature")
	errInvalidEventTypeFormat  = errors.New("invalid event type format")
	errMissingTimestamp        = errors.New("missing timestamp")
	errNoSignaturesFound       = errors.New("no signatures found")
	errTimestampTooOld         = errors.New("timestamp is too old")
	errTimestampTooFarInFuture = errors.New("timestamp is too far in the future")
	errInvalidTolerance        = errors.New("tolerance must be greater than 0")
)
