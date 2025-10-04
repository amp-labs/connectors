package calendly

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/go-playground/validator"
)

var (
	_ connectors.SubscribeConnector = &Connector{}
	
	ErrMissingOrganizationURI = errors.New("organization URI is required for webhook subscriptions")
	ErrMissingCallbackURL     = errors.New("callback URL is required for webhook subscriptions")
	ErrInvalidEventType       = errors.New("invalid event type")
	ErrSubscriptionFailed     = errors.New("subscription creation failed")
)

// CalendlySubscriptionRequest represents the request structure for creating webhook subscriptions
type CalendlySubscriptionRequest struct {
	URL      string                    `json:"url" validate:"required,url"`
	Events   []CalendlySubscriptionEvent `json:"events" validate:"required,min=1"`
	Organization string                `json:"organization" validate:"required"`
	User     string                    `json:"user,omitempty"`
	Scope    string                    `json:"scope" validate:"required,oneof=organization user"`
	SigningKey string                  `json:"signing_key,omitempty"`
}

// CalendlySubscriptionEvent represents the event types for Calendly webhooks
type CalendlySubscriptionEvent string

const (
	CalendlyEventInviteeCreated   CalendlySubscriptionEvent = "invitee.created"
	CalendlyEventInviteeCanceled  CalendlySubscriptionEvent = "invitee.canceled"
	CalendlyEventInviteeNoShowCreated CalendlySubscriptionEvent = "invitee_no_show.created"
	CalendlyEventInviteeNoShowDeleted CalendlySubscriptionEvent = "invitee_no_show.deleted"
	CalendlyEventRoutingFormSubmissionCreated CalendlySubscriptionEvent = "routing_form_submission.created"
)

// CalendlySubscriptionResult represents the response from creating/reading webhook subscriptions
type CalendlySubscriptionResult struct {
	URI                string                      `json:"uri"`
	CallbackURL        string                      `json:"callback_url"`
	SigningKey         string                      `json:"signing_key"`
	CreatedAt          string                      `json:"created_at"`
	UpdatedAt          string                      `json:"updated_at"`
	RetryStartedAt     *string                     `json:"retry_started_at"`
	State              string                      `json:"state"`
	Events             []CalendlySubscriptionEvent `json:"events"`
	Scope              string                      `json:"scope"`
	Organization       string                      `json:"organization"`
	User               *string                     `json:"user"`
	Group              *string                     `json:"group"`
	Creator            *string                     `json:"creator"`
}

// CalendlyWebhookSubscriptionsResponse represents the response from the webhook subscriptions list endpoint
type CalendlyWebhookSubscriptionsResponse struct {
	Collection []CalendlySubscriptionResult `json:"collection"`
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &CalendlySubscriptionRequest{},
	}
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &CalendlySubscriptionResult{},
	}
}

// Subscribe creates webhook subscriptions for specified events
func (c *Connector) Subscribe( //nolint:funlen
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, ok := params.Request.(*CalendlySubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type: expected *CalendlySubscriptionRequest, got %T", params.Request)
	}

	// Auto-populate organization and user URIs
	if req.Organization == "" {
		orgURI, err := c.getOrganizationURI(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get organization URI: %w", err)
		}
		req.Organization = orgURI
	}

	if req.Scope == "user" && req.User == "" {
		if c.userURI == "" {
			_, err := c.GetPostAuthInfo(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get user URI: %w", err)
			}
		}
		req.User = c.userURI
	}

	if err := c.validateSubscriptionParams(params); err != nil {
		return nil, err
	}

	result, err := c.createWebhookSubscription(ctx, req)
	if err != nil {
		return nil, err
	}

	subscriptionResult := &common.SubscriptionResult{
		Status: common.SubscriptionStatusSuccess,
		Result: result,
		Events: c.mapCalendlyEventsToCommon(req.Events),
	}

	return subscriptionResult, nil
}

// UpdateSubscription updates an existing webhook subscription
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// Calendly doesn't support direct updates, so delete and recreate
	if err := c.DeleteSubscription(ctx, *previousResult); err != nil {
		return nil, fmt.Errorf("failed to delete previous subscription: %w", err)
	}

	return c.Subscribe(ctx, params)
}

// DeleteSubscription removes a webhook subscription
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	previousResult common.SubscriptionResult,
) error {
	if previousResult.Result == nil {
		return fmt.Errorf("no subscription result to delete")
	}

	result, ok := previousResult.Result.(*CalendlySubscriptionResult)
	if !ok {
		return fmt.Errorf("invalid result type: expected *CalendlySubscriptionResult, got %T", previousResult.Result)
	}

	return c.deleteWebhookSubscription(ctx, result.URI)
}

// createWebhookSubscription creates a new webhook subscription via the Calendly API
func (c *Connector) createWebhookSubscription( //nolint:funlen
	ctx context.Context,
	req *CalendlySubscriptionRequest,
) (*CalendlySubscriptionResult, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "webhook_subscriptions")
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Post(ctx, url.String(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook subscription: %w", err)
	}

	body, ok := resp.Body()
	if !ok {
		return nil, fmt.Errorf("empty response body")
	}

	resource, err := jsonquery.New(body).ObjectRequired("resource")
	if err != nil {
		return nil, err
	}

	uri, err := jsonquery.New(resource).StringRequired("uri")
	if err != nil {
		return nil, err
	}

	callbackURL, err := jsonquery.New(resource).StringRequired("callback_url")
	if err != nil {
		return nil, err
	}

	createdAt, err := jsonquery.New(resource).StringRequired("created_at")
	if err != nil {
		return nil, err
	}

	updatedAt, err := jsonquery.New(resource).StringRequired("updated_at")
	if err != nil {
		return nil, err
	}

	state, err := jsonquery.New(resource).StringRequired("state")
	if err != nil {
		return nil, err
	}

	scope, err := jsonquery.New(resource).StringRequired("scope")
	if err != nil {
		return nil, err
	}

	organization, err := jsonquery.New(resource).StringRequired("organization")
	if err != nil {
		return nil, err
	}

	eventsData, err := jsonquery.New(resource).ArrayRequired("events")
	if err != nil {
		return nil, err
	}

	var events []CalendlySubscriptionEvent
	for _, eventNode := range eventsData {
		eventValue, err := eventNode.Value()
		if err != nil {
			continue
		}
		if eventStr, ok := eventValue.(string); ok {
			events = append(events, CalendlySubscriptionEvent(eventStr))
		}
	}

	// Optional fields
	var retryStartedAt *string
	if retryStr, err := jsonquery.New(resource).StringOptional("retry_started_at"); err == nil && retryStr != nil && *retryStr != "" {
		retryStartedAt = retryStr
	}

	var user *string
	if userStr, err := jsonquery.New(resource).StringOptional("user"); err == nil && userStr != nil && *userStr != "" {
		user = userStr
	}

	var group *string
	if groupStr, err := jsonquery.New(resource).StringOptional("group"); err == nil && groupStr != nil && *groupStr != "" {
		group = groupStr
	}

	var creator *string
	if creatorStr, err := jsonquery.New(resource).StringOptional("creator"); err == nil && creatorStr != nil && *creatorStr != "" {
		creator = creatorStr
	}

	result := &CalendlySubscriptionResult{
		URI:            uri,
		CallbackURL:    callbackURL,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		RetryStartedAt: retryStartedAt,
		State:          state,
		Events:         events,
		Scope:          scope,
		Organization:   organization,
		User:           user,
		Group:          group,
		Creator:        creator,
	}

	return result, nil
}

// deleteWebhookSubscription deletes a webhook subscription by URI
func (c *Connector) deleteWebhookSubscription(ctx context.Context, subscriptionURI string) error {
	subscriptionID, err := c.extractSubscriptionID(subscriptionURI)
	if err != nil {
		return err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "webhook_subscriptions", subscriptionID)
	if err != nil {
		return err
	}

	_, err = c.JSONHTTPClient().Delete(ctx, url.String())
	if err != nil {
		return fmt.Errorf("failed to delete webhook subscription: %w", err)
	}

	return nil
}

// validateSubscriptionParams validates the subscription parameters
func (c *Connector) validateSubscriptionParams(params common.SubscribeParams) error {
	if params.Request == nil {
		return fmt.Errorf("subscription request is required")
	}

	req, ok := params.Request.(*CalendlySubscriptionRequest)
	if !ok {
		return fmt.Errorf("invalid request type: expected *CalendlySubscriptionRequest, got %T", params.Request)
	}

	validator := validator.New()
	if err := validator.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate that events are supported
	for _, event := range req.Events {
		if !c.isValidCalendlyEvent(event) {
			return fmt.Errorf("%w: %s", ErrInvalidEventType, event)
		}
	}

	return nil
}

// isValidCalendlyEvent checks if the event type is supported
func (c *Connector) isValidCalendlyEvent(event CalendlySubscriptionEvent) bool {
	switch event {
	case CalendlyEventInviteeCreated, CalendlyEventInviteeCanceled,
		 CalendlyEventInviteeNoShowCreated, CalendlyEventInviteeNoShowDeleted,
		 CalendlyEventRoutingFormSubmissionCreated:
		return true
	default:
		return false
	}
}

// mapCalendlyEventsToCommon maps Calendly events to common subscription event types
func (c *Connector) mapCalendlyEventsToCommon(events []CalendlySubscriptionEvent) []common.SubscriptionEventType {
	var result []common.SubscriptionEventType
	
	for _, event := range events {
			switch event {
		case CalendlyEventInviteeCreated:
			result = append(result, common.SubscriptionEventTypeCreate)
		case CalendlyEventInviteeCanceled:
			result = append(result, common.SubscriptionEventTypeDelete)
		default:
			result = append(result, common.SubscriptionEventTypeOther)
		}
	}
	
	return result
}

// extractSubscriptionID extracts the subscription ID from a Calendly webhook subscription URI
func (c *Connector) extractSubscriptionID(uri string) (string, error) {
	if len(uri) == 0 {
		return "", fmt.Errorf("empty subscription URI")
	}

	lastSlash := -1
	for i := len(uri) - 1; i >= 0; i-- {
		if uri[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || lastSlash == len(uri)-1 {
		return "", fmt.Errorf("invalid subscription URI format: %s", uri)
	}

	return uri[lastSlash+1:], nil
}

// getOrganizationURI retrieves the organization URI from auth metadata or API
func (c *Connector) getOrganizationURI(ctx context.Context) (string, error) {
	authInfo, err := c.GetPostAuthInfo(ctx)
	if err == nil && authInfo != nil && authInfo.CatalogVars != nil {
		if orgURI, exists := (*authInfo.CatalogVars)["organizationURI"]; exists {
			return orgURI, nil
		}
	}

	// Retrieve from /users/me endpoint
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "users", "me")
	if err != nil {
		return "", err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	body, ok := resp.Body()
	if !ok {
		return "", fmt.Errorf("empty response body")
	}

	resource, err := jsonquery.New(body).ObjectRequired("resource")
	if err != nil {
		return "", err
	}

	orgURI, err := jsonquery.New(resource).StringRequired("current_organization")
	if err != nil {
		return "", fmt.Errorf("failed to get organization URI: %w", err)
	}

	return orgURI, nil
} 