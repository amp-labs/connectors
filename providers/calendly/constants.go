package calendly

// Normalized object names and Calendly API event-name prefixes (see webhook event strings).
const (
	objectNameEventTypes      = "event_types"
	objectNameScheduledEvents = "scheduled_events"
	objectNameRoutingForms    = "routing_forms"
)

const (
	calendlyPrefixEventType             = "event_type"
	calendlyPrefixInvitee               = "invitee"
	calendlyPrefixInviteeNoShow         = "invitee_no_show"
	calendlyPrefixRoutingFormSubmission = "routing_form_submission"
	// Object-name aliases accepted in SubscribeParams.SubscriptionEvents keys.
	objectAliasEventType   = "event_type"
	objectAliasRoutingForm = "routing_form"
)
