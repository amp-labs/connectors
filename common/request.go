// nolint:revive
package common

import (
	"context"
	"net/http"
)

type requestModifierKey string

type RequestModifier func(req *http.Request)

// WithRequestModifier adds a request modifier to the context. The request modifier
// will be called with the request before it is sent. This allows the caller to
// inject logic into the request, such as adding/removing headers or modifying the
// request body. Generally this should be used for debugging purposes only. Although
// there may be some use cases where it is useful to modify the request before it
// is sent. It can also double as an injectable observer function.
func WithRequestModifier(ctx context.Context, modifier RequestModifier) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, requestModifierKey("requestModifier"), modifier)
}

func getRequestModifier(ctx context.Context) (RequestModifier, bool) {
	if ctx == nil {
		return nil, false
	}

	value := ctx.Value(requestModifierKey("requestModifier"))
	if value == nil {
		return nil, false
	}

	modifier, ok := value.(RequestModifier)
	if !ok {
		return nil, false
	}

	if modifier == nil {
		return nil, false
	}

	return modifier, true
}
