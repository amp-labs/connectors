package acculynx

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/go-playground/validator"
)

// AccuLynx exposes a single subscription per installation. UpdateSubscription
// replaces the topic list in-place; consumerUrl is immutable after creation.
// Only contact_added/changed and job_created/updated are first-class; every
// other topic is reachable via ObjectEvents.PassThroughEvents.
//
// Reference: https://apidocs.acculynx.com/reference/postsubscription

const (
	webhooksVersionPrefix = "/webhooks/v2"
	subscriptionsPath     = "subscriptions"
)

var (
	errMissingSubscribeParams      = errors.New("acculynx: missing subscribe params")
	errInvalidSubscribeRequestType = errors.New("acculynx: invalid subscribe request type")
	errInvalidSubscriptionResult   = errors.New("acculynx: invalid subscription result type")
	errNoTopicsResolved            = errors.New("acculynx: no AccuLynx topics resolved from requested events")
	errUnsupportedSubscribeObject  = errors.New("acculynx: object does not support subscriptions")
	errUnsupportedSubscribeEvent   = errors.New("acculynx: event type not supported by AccuLynx for this object")
	errUnknownPassThroughTopic     = errors.New("acculynx: passThrough event is not a recognized AccuLynx topic")
	errEmptySubscriptionID         = errors.New("acculynx: empty subscriptionId in create response")
)

// SubscriptionRequest is the provider-specific input the framework passes via
// common.SubscribeParams.Request.
type SubscriptionRequest struct {
	// ConsumerURL is the HTTPS endpoint AccuLynx will POST events to. Immutable
	// after subscription creation.
	ConsumerURL string `json:"consumerUrl" validate:"required,url"`
	// TechContact is the email AccuLynx contacts about subscription status.
	TechContact string `json:"techContact" validate:"required,email"`
}

// SubscriptionResult is the persistent connector state stored between Subscribe
// calls. The framework keeps this verbatim and passes it back on UpdateSubscription
// / DeleteSubscription so we can identify the existing AccuLynx subscription.
type SubscriptionResult struct {
	SubscriptionID string   `json:"subscriptionId"`
	ConsumerURL    string   `json:"consumerUrl"`
	TechContact    string   `json:"techContact"`
	TopicNames     []string `json:"topicNames"`
	Status         string   `json:"status"`
}

// subscriptionCreateRequest mirrors the POST /webhooks/v2/subscriptions body.
type subscriptionCreateRequest struct {
	ConsumerURL     string   `json:"consumerUrl"`
	TechContact     string   `json:"techContact"`
	TopicNames      []string `json:"topicNames"`
	IntegrationType string   `json:"integrationType,omitempty"`
}

// subscriptionUpdateRequest mirrors the PUT /webhooks/v2/subscriptions/{id} body.
// Only technicalContact and topicNames are mutable on AccuLynx.
type subscriptionUpdateRequest struct {
	TechnicalContact string   `json:"technicalContact,omitempty"`
	TopicNames       []string `json:"topicNames"`
}

// subscriptionCreateResponse mirrors the POST response.
type subscriptionCreateResponse struct {
	SubscriptionID string `json:"subscriptionId"`
}

// objectEventTopics maps (object, framework event) to AccuLynx topic name.
// AccuLynx has no delete topics; non-standard topics go via PassThroughEvents.
//
//nolint:gochecknoglobals
var objectEventTopics = map[common.ObjectName]map[common.SubscriptionEventType][]string{
	objectContacts: {
		common.SubscriptionEventTypeCreate: {"contact_added"},
		common.SubscriptionEventTypeUpdate: {"contact_changed"},
	},
	objectJobs: {
		common.SubscriptionEventTypeCreate: {"job_created"},
		common.SubscriptionEventTypeUpdate: {"job_updated"},
	},
}

// validAcculynxTopics is the complete set of topicNames AccuLynx accepts on
// POST /webhooks/v2/subscriptions, verified live via GET /webhooks/v2/topics.
//
//nolint:gochecknoglobals
var validAcculynxTopics = datautils.NewStringSet(
	"contact_added",
	"contact_changed",
	"contact.custom-field.status_changed",
	"contact.custom-field.value_changed",
	"job_created",
	"job_updated",
	"job.accounting.integration-status.current_changed",
	"job.appointments.initial_created",
	"job.appointments.initial_updated",
	"job.category_changed",
	"job.contacts.primary_changed",
	"job.custom-field.status_changed",
	"job.custom-field.value_changed",
	"job.financials.approved-value_changed",
	"job.invoice_updated",
	"job.invoice_voided",
	"job.milestone.current_changed",
	"job.milestone.status.current_changed",
	"job.representatives.company_assigned",
	"job.representatives.company_changed",
	"job.trade-type_changed",
	"job.work-type_changed",
)

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

// Subscribe creates a single AccuLynx subscription containing every topic
// resolved from the requested SubscriptionEvents.
func (c *Connector) Subscribe(
	ctx context.Context, params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateSubscribeRequest(params)
	if err != nil {
		return nil, err
	}

	topics, err := resolveTopics(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	payload := subscriptionCreateRequest{
		ConsumerURL:     req.ConsumerURL,
		TechContact:     req.TechContact,
		TopicNames:      topics,
		IntegrationType: "Api",
	}

	created, err := c.createSubscription(ctx, payload)
	if err != nil {
		return nil, err
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
		Result: &SubscriptionResult{
			SubscriptionID: created.SubscriptionID,
			ConsumerURL:    req.ConsumerURL,
			TechContact:    req.TechContact,
			TopicNames:     topics,
			Status:         "enabled",
		},
	}, nil
}

// UpdateSubscription replaces the topic list on the existing subscription.
// Returns an error if the caller asks to change consumerUrl (immutable on AccuLynx).
func (c *Connector) UpdateSubscription(
	ctx context.Context, params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	req, err := validateSubscribeRequest(params)
	if err != nil {
		return nil, err
	}

	prev, err := extractSubscriptionResult(previousResult)
	if err != nil {
		return nil, err
	}

	if req.ConsumerURL != prev.ConsumerURL {
		return nil, fmt.Errorf("%w: consumerUrl is immutable on AccuLynx (%s -> %s)",
			errInvalidSubscribeRequestType, prev.ConsumerURL, req.ConsumerURL)
	}

	topics, err := resolveTopics(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	payload := subscriptionUpdateRequest{
		TechnicalContact: req.TechContact,
		TopicNames:       topics,
	}

	if err := c.updateSubscription(ctx, prev.SubscriptionID, payload); err != nil {
		return nil, err
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
		Result: &SubscriptionResult{
			SubscriptionID: prev.SubscriptionID,
			ConsumerURL:    prev.ConsumerURL,
			TechContact:    req.TechContact,
			TopicNames:     topics,
			Status:         prev.Status,
		},
	}, nil
}

// DeleteSubscription removes the subscription identified by previousResult.
func (c *Connector) DeleteSubscription(
	ctx context.Context, previousResult common.SubscriptionResult,
) error {
	prev, err := extractSubscriptionResult(&previousResult)
	if err != nil {
		return err
	}

	return c.deleteSubscription(ctx, prev.SubscriptionID)
}

func (c *Connector) createSubscription(
	ctx context.Context, payload subscriptionCreateRequest,
) (*subscriptionCreateResponse, error) {
	u, err := c.subscriptionsURL()
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Post(ctx, u.String(), payload)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	parsed, err := common.UnmarshalJSON[subscriptionCreateResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("parse create subscription response: %w", err)
	}

	if parsed == nil || parsed.SubscriptionID == "" {
		return nil, errEmptySubscriptionID
	}

	return parsed, nil
}

func (c *Connector) updateSubscription(
	ctx context.Context, subscriptionID string, payload subscriptionUpdateRequest,
) error {
	u, err := c.subscriptionByIDURL(subscriptionID)
	if err != nil {
		return err
	}

	_, err = c.JSONHTTPClient().Put(ctx, u.String(), payload)
	if err != nil {
		return fmt.Errorf("update subscription %s: %w", subscriptionID, err)
	}

	return nil
}

func (c *Connector) deleteSubscription(ctx context.Context, subscriptionID string) error {
	u, err := c.subscriptionByIDURL(subscriptionID)
	if err != nil {
		return err
	}

	_, err = c.JSONHTTPClient().Delete(ctx, u.String())
	if err != nil {
		return fmt.Errorf("delete subscription %s: %w", subscriptionID, err)
	}

	return nil
}

func (c *Connector) subscriptionsURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, webhooksVersionPrefix, subscriptionsPath)
}

func (c *Connector) subscriptionByIDURL(subscriptionID string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, webhooksVersionPrefix, subscriptionsPath, subscriptionID)
}

func validateSubscribeRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: Request is nil", errMissingSubscribeParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected *SubscriptionRequest, got %T",
			errInvalidSubscribeRequestType, params.Request)
	}

	if err := validator.New().Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidSubscribeRequestType, err)
	}

	return req, nil
}

func extractSubscriptionResult(result *common.SubscriptionResult) (*SubscriptionResult, error) {
	if result == nil || result.Result == nil {
		return nil, fmt.Errorf("%w: previousResult.Result is nil", errInvalidSubscriptionResult)
	}

	prev, ok := result.Result.(*SubscriptionResult)
	if !ok {
		return nil, fmt.Errorf("%w: expected *SubscriptionResult, got %T",
			errInvalidSubscriptionResult, result.Result)
	}

	if prev.SubscriptionID == "" {
		return nil, fmt.Errorf("%w: missing SubscriptionID", errInvalidSubscriptionResult)
	}

	return prev, nil
}

// resolveTopics returns the deduplicated, sorted AccuLynx topic-name slice
// for the given events. Objects must be in objectEventTopics; pass-through
// events must be in validAcculynxTopics.
func resolveTopics(events map[common.ObjectName]common.ObjectEvents) ([]string, error) {
	seen := make(map[string]struct{})

	for obj, objEvents := range events {
		if err := appendObjectTopics(obj, objEvents, seen); err != nil {
			return nil, err
		}
	}

	if len(seen) == 0 {
		return nil, errNoTopicsResolved
	}

	topics := make([]string, 0, len(seen))
	for t := range seen {
		topics = append(topics, t)
	}

	slices.Sort(topics)

	return topics, nil
}

func appendObjectTopics(
	obj common.ObjectName,
	objEvents common.ObjectEvents,
	seen map[string]struct{},
) error {
	objMap, supported := objectEventTopics[obj]
	if !supported && len(objEvents.PassThroughEvents) == 0 {
		return fmt.Errorf("%w: %s", errUnsupportedSubscribeObject, obj)
	}

	for _, evt := range objEvents.Events {
		topics, ok := objMap[evt]
		if !ok {
			return fmt.Errorf("%w: object=%s event=%s",
				errUnsupportedSubscribeEvent, obj, evt)
		}

		for _, t := range topics {
			seen[t] = struct{}{}
		}
	}

	for _, raw := range objEvents.PassThroughEvents {
		if !validAcculynxTopics.Has(raw) {
			return fmt.Errorf("%w: %s", errUnknownPassThroughTopic, raw)
		}

		seen[raw] = struct{}{}
	}

	return nil
}
