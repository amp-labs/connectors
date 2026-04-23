package housecallpro

import "errors"

var (
	errInvalidEventTypeFormat     = errors.New("invalid webhook event type format")
	errInvalidSignature           = errors.New("invalid webhook signature")
	errMalformedWebhookEvent      = errors.New("malformed webhook event")
	errMissingParams              = errors.New("missing required parameters")
	errUnsupportedWebhookResource = errors.New("webhook event resource is not supported")
)
