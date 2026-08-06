package stripe

import "errors"

var (
	errMissingParams           = errors.New("missing required parameters")
	errMissingSignature        = errors.New("missing webhook signature header")
	errInvalidSignature        = errors.New("invalid webhook signature")
	errInvalidEventTypeFormat  = errors.New("invalid event type format")
	errMissingTimestamp        = errors.New("missing timestamp")
	errNoSignaturesFound       = errors.New("no signatures found")
	errTimestampTooOld         = errors.New("timestamp is too old")
	errTimestampTooFarInFuture = errors.New("timestamp is too far in the future")
	errInvalidTolerance        = errors.New("tolerance must be greater than 0")
)
