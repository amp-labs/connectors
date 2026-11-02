package reply

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/go-playground/validator"
)

var _ connectors.SubscribeConnector = &Connector{}

// A Reply webhook subscription delivers a single event type to a URL, so subscribing to
// several events creates one subscription per event. Reply needs no setup beforehand and has
// no shared resource the subscriptions hang off, which is why there is no registration step.
//
// Every event is an activity, never a create, update or delete of a record, so events are
// requested by their Reply name through ObjectEvents.PassThroughEvents (see eventRoutes).
//
// Reply accepts a second subscription for the same event and URL, so a Subscribe retried after
// the first one's result was lost (a crash between Reply creating the subscriptions and the
// caller storing the result) subscribes twice and every event is delivered twice. Like the
// other providers that track subscriptions only in the stored result, this does not list the
// subscriptions Reply already holds to find them.
//
// https://docs.reply.io/api-reference/webhooks/create-a-webhook-subscription

const (
	webhooksPath  = "webhooks"
	scopePersonal = "personal"
)

var (
	errMissingSubscribeParams    = errors.New("reply subscribe: missing subscribe params")
	errInvalidSubscribeRequest   = errors.New("reply subscribe: invalid subscribe request")
	errInvalidSubscriptionResult = errors.New("reply subscribe: invalid subscription result")
	errNoEventsRequested         = errors.New("reply subscribe: no events requested")
	errCRUDEventsNotSupported    = errors.New(
		"reply subscribe: Reply sends no create, update or delete events, request its events as pass-through events",
	)
	errUnknownPassThroughEvent = errors.New("reply subscribe: not an event Reply offers")
	errEventObjectMismatch     = errors.New("reply subscribe: event does not belong to the object")
	errEmptySubscriptionID     = errors.New("reply subscribe: empty id in create response")
)

// SubscriptionRequest is the provider-specific input the caller passes in
// common.SubscribeParams.Request.
type SubscriptionRequest struct {
	// WebhookURL is the endpoint Reply will POST every event to. Reply takes up to 1024 characters.
	WebhookURL string `json:"webhookUrl" validate:"required,url,max=1024"`
}

// SubscriptionResult is the persistent connector state: every Reply subscription that exists
// for the installation. The caller stores it and hands it back to update and delete.
type SubscriptionResult struct {
	Subscriptions []WebhookSubscription `json:"subscriptions"`
}

// WebhookSubscription is one Reply webhook subscription.
type WebhookSubscription struct {
	ID        int    `json:"id"`
	EventType string `json:"eventType"`
	URL       string `json:"url"`
	Scope     string `json:"scope"`
}

// webhookPayload mirrors the POST /v3/webhooks body. The scope is always personal, and the
// payload flags are left out so Reply uses its defaults (all off).
type webhookPayload struct {
	EventType string `json:"eventType"`
	URL       string `json:"url"`
	Scope     string `json:"scope"`
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &SubscriptionRequest{},
	}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &SubscriptionResult{},
	}
}

// Subscribe creates one Reply subscription per requested event. If any creation fails the ones
// already created are deleted again, so Reply is left as it was found.
func (c *Connector) Subscribe(
	ctx context.Context, params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	request, err := validateSubscribeRequest(params)
	if err != nil {
		return nil, err
	}

	events, err := resolveEvents(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	created, createErr := c.createSubscriptions(ctx, request.WebhookURL, events)
	if createErr != nil {
		leftover, rollbackErr := c.deleteSubscriptions(ctx, created)
		if rollbackErr != nil {
			return subscriptionResult(leftover, common.SubscriptionStatusFailedToRollback),
				errors.Join(createErr, rollbackErr)
		}

		return &common.SubscriptionResult{Status: common.SubscriptionStatusFailed}, createErr
	}

	return subscriptionResult(created, common.SubscriptionStatusSuccess), nil
}

// UpdateSubscription reconciles the existing subscriptions with the requested events. Events
// that are no longer wanted are deleted, new ones are created, and an event whose URL changed
// is created again at the new URL. New subscriptions are created first so no event is missed,
// and are rolled back if that fails, leaving the previous state intact.
func (c *Connector) UpdateSubscription(
	ctx context.Context, params common.SubscribeParams, previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	request, err := validateSubscribeRequest(params)
	if err != nil {
		return nil, err
	}

	events, err := resolveEvents(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	previous, err := extractSubscriptionResult(previousResult)
	if err != nil {
		return nil, err
	}

	keep, stale, missing := reconcile(previous.Subscriptions, events, request.WebhookURL)

	created, createErr := c.createSubscriptions(ctx, request.WebhookURL, missing)
	if createErr != nil {
		leftover, rollbackErr := c.deleteSubscriptions(ctx, created)
		if rollbackErr != nil {
			// What could not be rolled back still exists next to everything that was there before.
			state := append(slices.Clone(previous.Subscriptions), leftover...)

			return subscriptionResult(state, common.SubscriptionStatusFailedToRollback),
				errors.Join(createErr, rollbackErr)
		}

		return subscriptionResult(previous.Subscriptions, common.SubscriptionStatusFailed), createErr
	}

	undeleted, deleteErr := c.deleteSubscriptions(ctx, stale)

	state := slices.Concat(keep, created, undeleted)

	if deleteErr != nil {
		return subscriptionResult(state, common.SubscriptionStatusFailedToRollback), deleteErr
	}

	return subscriptionResult(state, common.SubscriptionStatusSuccess), nil
}

// DeleteSubscription deletes every subscription in the previous result. A subscription that is
// already gone counts as deleted. It keeps going after a failure and reports all of them.
func (c *Connector) DeleteSubscription(ctx context.Context, previousResult common.SubscriptionResult) error {
	previous, err := extractSubscriptionResult(&previousResult)
	if err != nil {
		return err
	}

	_, err = c.deleteSubscriptions(ctx, previous.Subscriptions)

	return err
}

// createSubscriptions creates one subscription per event, in order, and stops at the first
// failure. It returns what was created, including on failure, so the caller can roll back.
// Creation is sequential: there are at most 22 events and Reply allows 100 requests a minute.
func (c *Connector) createSubscriptions(
	ctx context.Context, webhookURL string, events []string,
) ([]WebhookSubscription, error) {
	created := make([]WebhookSubscription, 0, len(events))

	for _, event := range events {
		subscription, err := c.createSubscription(ctx, webhookURL, event)
		if err != nil {
			return created, fmt.Errorf("subscribe to %s: %w", event, err)
		}

		created = append(created, *subscription)
	}

	return created, nil
}

func (c *Connector) createSubscription(
	ctx context.Context, webhookURL, event string,
) (*WebhookSubscription, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, webhooksPath)
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Post(ctx, url.String(), webhookPayload{
		EventType: event,
		URL:       webhookURL,
		Scope:     scopePersonal,
	})
	if err != nil {
		return nil, err
	}

	subscription, err := common.UnmarshalJSON[WebhookSubscription](response)
	if err != nil {
		return nil, err
	}

	if subscription == nil || subscription.ID == 0 {
		return nil, errEmptySubscriptionID
	}

	return subscription, nil
}

// deleteSubscriptions deletes each subscription and returns the ones that could not be
// deleted, with the errors. A 404 means the subscription is already gone.
func (c *Connector) deleteSubscriptions(
	ctx context.Context, subscriptions []WebhookSubscription,
) ([]WebhookSubscription, error) {
	var (
		remaining []WebhookSubscription
		failures  []error
	)

	for _, subscription := range subscriptions {
		err := c.deleteSubscription(ctx, subscription.ID)
		if err != nil && !isNotFound(err) {
			remaining = append(remaining, subscription)
			failures = append(failures,
				fmt.Errorf("delete subscription %d (%s): %w", subscription.ID, subscription.EventType, err))
		}
	}

	return remaining, errors.Join(failures...)
}

func (c *Connector) deleteSubscription(ctx context.Context, id int) error {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, webhooksPath, strconv.Itoa(id))
	if err != nil {
		return err
	}

	_, err = c.JSONHTTPClient().Delete(ctx, url.String())

	return err
}

// reconcile splits the existing subscriptions against the wanted events: the ones to keep,
// the stale ones to delete (no longer wanted, or pointing at another URL), and the events
// that still need a subscription.
func reconcile(
	existing []WebhookSubscription, wanted []string, webhookURL string,
) (keep, stale []WebhookSubscription, missing []string) {
	wantedSet := datautils.NewSetFromList(wanted)
	covered := datautils.NewSet[string]()

	for _, subscription := range existing {
		if wantedSet.Has(subscription.EventType) && subscription.URL == webhookURL && !covered.Has(subscription.EventType) {
			keep = append(keep, subscription)
			covered.AddOne(subscription.EventType)

			continue
		}

		stale = append(stale, subscription)
	}

	for _, event := range wanted {
		if !covered.Has(event) {
			missing = append(missing, event)
		}
	}

	return keep, stale, missing
}

// resolveEvents returns the sorted, de-duplicated Reply event names to subscribe to. Every
// event must be one Reply offers and be requested under the object it concerns.
func resolveEvents(requested map[common.ObjectName]common.ObjectEvents) ([]string, error) {
	events := datautils.NewSet[string]()

	for object, objectEvents := range requested {
		if len(objectEvents.Events) > 0 {
			return nil, fmt.Errorf("%w: %s", errCRUDEventsNotSupported, object)
		}

		for _, event := range objectEvents.PassThroughEvents {
			route, ok := eventRoutes[event]
			if !ok {
				return nil, fmt.Errorf("%w: %s", errUnknownPassThroughEvent, event)
			}

			if route.object != string(object) {
				return nil, fmt.Errorf("%w: %s belongs to %s, not %s",
					errEventObjectMismatch, event, route.object, object)
			}

			events.AddOne(event)
		}
	}

	if events.IsEmpty() {
		return nil, errNoEventsRequested
	}

	list := events.List()
	slices.Sort(list)

	return list, nil
}

// subscriptionResult builds the result for the given subscriptions: the stored state, and
// the objects and events it covers.
func subscriptionResult(
	subscriptions []WebhookSubscription, status common.SubscriptionStatus,
) *common.SubscriptionResult {
	sorted := slices.Clone(subscriptions)
	slices.SortFunc(sorted, func(a, b WebhookSubscription) int {
		if byEvent := cmp.Compare(a.EventType, b.EventType); byEvent != 0 {
			return byEvent
		}

		return a.ID - b.ID
	})

	objectEvents := make(map[common.ObjectName]common.ObjectEvents)

	for _, subscription := range sorted {
		route, ok := eventRoutes[subscription.EventType]
		if !ok {
			continue
		}

		object := common.ObjectName(route.object)
		current := objectEvents[object]

		if !slices.Contains(current.PassThroughEvents, subscription.EventType) {
			current.PassThroughEvents = append(current.PassThroughEvents, subscription.EventType)
		}

		objectEvents[object] = current
	}

	return &common.SubscriptionResult{
		Status:       status,
		Result:       &SubscriptionResult{Subscriptions: sorted},
		ObjectEvents: objectEvents,
	}
}

func validateSubscribeRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: Request is nil", errMissingSubscribeParams)
	}

	request, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected *SubscriptionRequest, got %T", errInvalidSubscribeRequest, params.Request)
	}

	if err := validator.New().Struct(request); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidSubscribeRequest, err)
	}

	return request, nil
}

func extractSubscriptionResult(result *common.SubscriptionResult) (*SubscriptionResult, error) {
	if result == nil || result.Result == nil {
		return nil, fmt.Errorf("%w: previous result is nil", errInvalidSubscriptionResult)
	}

	previous, ok := result.Result.(*SubscriptionResult)
	if !ok {
		return nil, fmt.Errorf("%w: expected *SubscriptionResult, got %T", errInvalidSubscriptionResult, result.Result)
	}

	return previous, nil
}
