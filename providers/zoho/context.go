package zoho

import "context"

type locationKey string

const locationCtxKey locationKey = "location"

// WithLocation adds location to the context. This comes from a query parameter
// in the OAuth callback URL. It is required for Zoho CRM API calls.
func WithLocation(ctx context.Context, location string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, locationCtxKey, location)
}

// getLocation retrieves the location from the context.
func getLocation(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	val := ctx.Value(locationCtxKey)
	if val == nil {
		return "", false
	}

	location, ok := val.(string)

	return location, ok
}
