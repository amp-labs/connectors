package providers

// customUnauthorizedDeciders holds per-provider handlers that decide whether a
// response indicates an expired or invalid session, for custom-auth providers
// whose APIs don't signal it with a plain HTTP 401. Populated in provider init()
// via RegisterCustomUnauthorizedDecider and read when the custom-auth client is
// built, so the reactive re-auth path works without the caller wiring it.
var customUnauthorizedDeciders = map[Provider]IsUnauthorizedDecider{} //nolint:gochecknoglobals

// RegisterCustomUnauthorizedDecider records the unauthorized decider for a
// provider whose API reports an expired session with something other than a 401
// (e.g. Bill.com returns HTTP 200 with a non-zero response_status).
func RegisterCustomUnauthorizedDecider(provider Provider, decider IsUnauthorizedDecider) {
	customUnauthorizedDeciders[provider] = decider
}

// CustomUnauthorizedDeciderFor returns the registered decider for a provider, or
// nil if the provider uses the default (HTTP 401) unauthorized signal.
func CustomUnauthorizedDeciderFor(provider Provider) IsUnauthorizedDecider {
	return customUnauthorizedDeciders[provider]
}
