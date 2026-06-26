package calendly

import "errors"

var (
	errUnexpectedEventName         = errors.New("calendly webhook: unexpected event name")
	errUnsupportedWebhookFamily    = errors.New("calendly webhook: unsupported webhook event family")
	errWebhookOrgNotFound          = errors.New("calendly webhook: organization not found")
	errWebhookRecordURINotFound    = errors.New("calendly webhook: record uri not found in payload")
	errWebhookTimestampNotFound    = errors.New("calendly webhook: timestamp not found in payload")
	errWebhookPayloadNotObject     = errors.New("calendly webhook: payload is not an object")
	errCalendlySigningKeyEmpty     = errors.New("calendly: signing key is empty")
	errCalendlyMissingSigHeader    = errors.New("calendly: missing Calendly-Webhook-Signature header")
	errCalendlySigHeaderFormat     = errors.New("calendly: signature header missing t= or v1=")
	errCalendlyEmptySignatureBytes = errors.New("calendly: empty signature bytes")
)
