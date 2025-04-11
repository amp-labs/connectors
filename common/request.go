package common

import (
	"context"
	"net/http"
)

type requestModifierKey string

type RequestModifier func(req *http.Request)

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
