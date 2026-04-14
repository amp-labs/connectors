package calendly

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/go-playground/validator"
)

// Compile-time interface checks.
var _ connectors.SubscribeConnector = (*Connector)(nil)

var (
	errMissingSubscriptionParams  = errors.New("missing required subscription parameters")
	errInvalidSubscriptionRequest = errors.New("invalid subscription request type")
)

// SubscriptionRequest is the Calendly-specific payload for common.SubscribeParams.Request.
// It maps to POST https://api.calendly.com/webhook_subscriptions (see Calendly webhook docs).
type SubscriptionRequest struct {
	// CallbackURL is the webhook destination URL (JSON field "url").
	CallbackURL string `json:"url" validate:"required"`
	// SigningKey is used by Calendly to sign outbound webhooks ("signing_key").
	SigningKey string `json:"signing_key" validate:"required"`
	// Scope is one of: organization, user, group.
	Scope string `json:"scope" validate:"required,oneof=organization user group"`
	// OrganizationURI optional override; defaults to catalog org URI from post-auth.
	OrganizationURI string `json:"organization,omitempty"`
	// UserURI optional override; defaults to catalog user URI from post-auth.
	UserURI string `json:"user,omitempty"`
	// GroupURI required when Scope is "group".
	GroupURI string `json:"group,omitempty"`
	// Events optional explicit Calendly event names (e.g. event_type.created). Merged with
	// events derived from SubscriptionEvents when both are present.
	Events []string `json:"events,omitempty"`
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &SubscriptionRequest{},
	}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: map[string]any{},
	}
}

// Subscribe creates a single Calendly webhook subscription for the requested events.
func (c *Connector) Subscribe(ctx context.Context, params common.SubscribeParams) (*common.SubscriptionResult, error) {
	req, err := c.parseSubscriptionRequest(params)
	if err != nil {
		return nil, err
	}

	if err := c.validateScopeURIs(req); err != nil {
		return nil, err
	}

	events := buildCalendlyEventList(params, req)
	if len(events) == 0 {
		return nil, fmt.Errorf("%w: no events to subscribe (set SubscriptionEvents and/or SubscriptionRequest.Events)",
			errMissingSubscriptionParams)
	}

	body := c.buildWebhookSubscriptionBody(req, events)

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "webhook_subscriptions")
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Post(ctx, url.String(), body)
	if err != nil {
		return nil, err
	}

	parsed, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	objectEvents := objectEventsForResult(params, events)

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		Result:       *parsed,
		ObjectEvents: objectEvents,
	}, nil
}

// UpdateSubscription replaces an existing webhook subscription by deleting it and creating a new one.
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	if previousResult == nil || previousResult.Result == nil {
		return nil, fmt.Errorf("%w: previous subscription result is required", errMissingSubscriptionParams)
	}

	uri, err := extractSubscriptionURI(previousResult.Result)
	if err != nil {
		return nil, err
	}

	if _, err := c.JSONHTTPClient().Delete(ctx, uri); err != nil {
		return nil, fmt.Errorf("calendly: delete previous webhook subscription: %w", err)
	}

	return c.Subscribe(ctx, params)
}

// DeleteSubscription removes a webhook subscription using the URI returned from Subscribe/Update.
func (c *Connector) DeleteSubscription(ctx context.Context, previousResult common.SubscriptionResult) error {
	uri, err := extractSubscriptionURI(previousResult.Result)
	if err != nil {
		return err
	}

	_, err = c.JSONHTTPClient().Delete(ctx, uri)

	return err
}

func (c *Connector) parseSubscriptionRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: SubscribeParams.Request is nil", errMissingSubscriptionParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected *calendly.SubscriptionRequest, got %T",
			errInvalidSubscriptionRequest, params.Request)
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("calendly subscription request: %w", err)
	}

	return req, nil
}

func (c *Connector) validateScopeURIs(req *SubscriptionRequest) error {
	org := firstNonEmpty(req.OrganizationURI, c.orgURI)
	user := firstNonEmpty(req.UserURI, c.userURI)

	switch req.Scope {
	case "organization":
		if org == "" {
			return fmt.Errorf("%w: organization URI required (metadata or SubscriptionRequest.organization)",
				errMissingSubscriptionParams)
		}
	case "user":
		if user == "" {
			return fmt.Errorf("%w: user URI required (metadata or SubscriptionRequest.user)",
				errMissingSubscriptionParams)
		}

		if org == "" {
			return fmt.Errorf("%w: organization URI required for user scope",
				errMissingSubscriptionParams)
		}
	case "group":
		if req.GroupURI == "" {
			return fmt.Errorf("%w: group URI required for group scope", errMissingSubscriptionParams)
		}
	}

	return nil
}

func (c *Connector) buildWebhookSubscriptionBody(req *SubscriptionRequest, events []string) map[string]any {
	org := firstNonEmpty(req.OrganizationURI, c.orgURI)
	user := firstNonEmpty(req.UserURI, c.userURI)

	body := map[string]any{
		"url":         req.CallbackURL,
		"events":      events,
		"scope":       req.Scope,
		"signing_key": req.SigningKey,
	}

	if org != "" {
		body["organization"] = org
	}

	if user != "" {
		body["user"] = user
	}

	if req.Scope == "group" && req.GroupURI != "" {
		body["group"] = req.GroupURI
	}

	return body
}

func buildCalendlyEventList(params common.SubscribeParams, req *SubscriptionRequest) []string {
	out := append([]string(nil), req.Events...)

	for objName, oe := range params.SubscriptionEvents {
		prefix := calendlyEventPrefix(string(objName))
		if prefix != "" {
			for _, ev := range oe.Events {
				if name, ok := calendlyEventFromCRUD(prefix, ev); ok {
					out = append(out, name)
				}
			}
		}

		out = append(out, oe.PassThroughEvents...)
	}

	return datautils.NewSetFromList(out).List()
}

func calendlyEventPrefix(objectName string) string {
	switch objectName {
	case objectNameEventTypes, objectAliasEventType:
		return calendlyPrefixEventType
	case objectNameScheduledEvents, calendlyPrefixInvitee:
		return calendlyPrefixInvitee
	case objectNameRoutingForms, objectAliasRoutingForm:
		return calendlyPrefixRoutingFormSubmission
	default:
		return ""
	}
}

func calendlyEventFromCRUD(prefix string, ev common.SubscriptionEventType) (string, bool) {
	switch ev {
	case common.SubscriptionEventTypeCreate:
		return prefix + ".created", true
	case common.SubscriptionEventTypeUpdate:
		return prefix + ".updated", true
	case common.SubscriptionEventTypeDelete:
		return prefix + ".deleted", true
	case common.SubscriptionEventTypeAssociationUpdate, common.SubscriptionEventTypeOther:
		return "", false
	default:
		return "", false
	}
}

func objectEventsForResult(
	params common.SubscribeParams,
	resolvedEvents []string,
) map[common.ObjectName]common.ObjectEvents {
	if len(params.SubscriptionEvents) > 0 {
		return maps.Clone(params.SubscriptionEvents)
	}

	return inferObjectEventsFromCalendlyEvents(resolvedEvents)
}

func inferObjectEventsFromCalendlyEvents(events []string) map[common.ObjectName]common.ObjectEvents {
	out := make(map[common.ObjectName]common.ObjectEvents)

	for _, ev := range events {
		objName, subEv := inferObjectAndNormalizedEvent(ev)
		if objName == "" {
			continue
		}

		cur := out[objName]
		cur.Events = append(cur.Events, subEv)
		out[objName] = cur
	}

	for k, v := range out {
		v.Events = datautils.NewSetFromList(v.Events).List()
		out[k] = v
	}

	return out
}

func objectNameFromCalendlyPrefix(prefix string) common.ObjectName {
	switch prefix {
	case calendlyPrefixEventType:
		return objectNameEventTypes
	case calendlyPrefixInvitee, calendlyPrefixInviteeNoShow:
		return objectNameScheduledEvents
	case calendlyPrefixRoutingFormSubmission:
		return objectNameRoutingForms
	default:
		return ""
	}
}

func normalizedEventFromAction(action string) (common.SubscriptionEventType, bool) {
	switch action {
	case "created":
		return common.SubscriptionEventTypeCreate, true
	case "updated":
		return common.SubscriptionEventTypeUpdate, true
	case "deleted", "canceled":
		return common.SubscriptionEventTypeDelete, true
	default:
		return "", false
	}
}

func inferObjectAndNormalizedEvent(calendlyEvent string) (common.ObjectName, common.SubscriptionEventType) {
	parts := splitEventName(calendlyEvent)
	if len(parts) != 2 { //nolint:mnd
		return "", ""
	}

	prefix, action := parts[0], parts[1]

	obj := objectNameFromCalendlyPrefix(prefix)
	if obj == "" {
		return "", ""
	}

	eventType, ok := normalizedEventFromAction(action)
	if !ok {
		return "", ""
	}

	return obj, eventType
}

func splitEventName(calendlyEvent string) []string {
	// Handles "routing_form_submission.created" by splitting on the last dot for the action,
	// and the remainder as prefix with internal dots.
	idx := -1
	for i := len(calendlyEvent) - 1; i >= 0; i-- {
		if calendlyEvent[i] == '.' {
			idx = i

			break
		}
	}

	if idx <= 0 {
		return nil
	}

	return []string{calendlyEvent[:idx], calendlyEvent[idx+1:]}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}

	return b
}

func extractSubscriptionURI(result any) (string, error) {
	switch v := result.(type) {
	case map[string]any:
		if r, ok := v["resource"].(map[string]any); ok {
			if uri, ok := r["uri"].(string); ok && uri != "" {
				return uri, nil
			}
		}
	case *map[string]any:
		return extractSubscriptionURI(*v)
	}

	return "", fmt.Errorf("%w: subscription URI not found in result", errMissingSubscriptionParams)
}
