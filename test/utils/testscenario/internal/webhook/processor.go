package webhook

import (
	"context"
	"fmt"
	"net/http"
)

// Processor handles verified webhook events and optional provider-specific
// request interception.
//
// Interceptor can short-circuit special requests such as validation
// challenges. The internal channel carries verified events to Processor.Run.
type Processor struct {
	// Interceptor may handle a request before it is connector-verified.
	// If it returns true, the request is considered handled and no further
	// processing occurs.
	Interceptor InterceptorFunc

	// channel is managed internally and receives verified events from Server.
	// Processor.Run consumes events from this channel using an OnEventFunc
	// callback. It is intended to be driven by the test scripts.
	channel chan Event
}

// InterceptorFunc allows custom handling of incoming HTTP requests before they are verified.
// If it returns true, the request is considered handled and no further processing occurs.
type InterceptorFunc func(writer http.ResponseWriter, request *http.Request, requestBody []byte) bool

// OnEventFunc handles a verified webhook event received by Processor.
//
// It returns true when the caller should stop processing more events, and
// false when processing should continue.
type OnEventFunc func(Event) bool

// processEvent sends a WebhookEvent to the channel if possible,
// or discards it if the context is canceled or done.
//
// This is useful when the provider may still be sending events while
// the server is shutting down (for example, because the expected number
// of webhook events has arrived, or because a developer canceled the test script).
// In that case, the send will not block the handler; instead, the message is silently dropped.
func (p Processor) processEvent(ctx context.Context, event Event) {
	select {
	case p.channel <- event:
	case <-ctx.Done():
		// We don't care if we couldn't send.
		// The server is shutting down.
	}
}

func (p Processor) Run(ctx context.Context, onEvent OnEventFunc) {
	for {
		select {
		case msg := <-p.channel:
			if onEvent == nil {
				fmt.Println("OnEvent is nil, received a message, don't know what to do.")
				return
			}

			if done := onEvent(msg); done {
				return
			}
		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping...")
			return
		}
	}
}
