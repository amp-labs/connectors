package jobber

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/go-playground/validator"
)

var _ connectors.SubscribeConnector = &Connector{}

// Jobber delivers webhooks per topic ("{OBJECT}_{EVENT}", e.g. CLIENT_CREATE).
// Endpoints are managed with the webhookEndpointCreate / webhookEndpointDelete
// GraphQL mutations, one endpoint per topic, all pointing at the same URL.
// Jobber has no update or list operation for webhook endpoints, so
// UpdateSubscription reconciles by deleting stale endpoints and creating new
// ones, and DeleteSubscription relies on the endpoint IDs stored at creation.
//
// Reference: https://developer.getjobber.com/docs/using_jobbers_api/setting_up_webhooks/

const maxConcurrentSubscriptions = 4

var (
	errMissingSubscribeParams      = errors.New("jobber: missing subscribe params")
	errInvalidSubscribeRequest     = errors.New("jobber: invalid subscribe request")
	errNoTopicsResolved            = errors.New("jobber: no webhook topics resolved from requested events")
	errUnsupportedSubscribeObject  = errors.New("jobber: object does not support subscriptions")
	errUnsupportedSubscribeEvent   = errors.New("jobber: event type not supported by Jobber for this object")
	errUnknownPassThroughTopic     = errors.New("jobber: passThrough event is not a recognized Jobber webhook topic")
	errEmptyWebhookEndpoint        = errors.New("jobber: empty webhookEndpoint in create response")
	errSubscriptionUserErrors      = errors.New("jobber: provider returned userErrors")
	errMissingStoredEndpoints      = errors.New("jobber: previous subscription result has no endpoints")
	errFailedToCreateSubscriptions = errors.New("jobber: failed to create webhook endpoints")
)

// SubscriptionRequest is the provider-specific input the framework passes via
// common.SubscribeParams.Request.
type SubscriptionRequest struct {
	// WebhookURL is the HTTPS endpoint Jobber will POST webhook events to.
	WebhookURL string `json:"webhookUrl" validate:"required,url"`
}

// WebhookEndpoint is one Jobber webhook endpoint (one topic -> one URL).
type WebhookEndpoint struct {
	ID    string `json:"id"`
	Topic string `json:"topic"`
	URL   string `json:"url"`
}

// SubscriptionResult is the persistent connector state stored between
// Subscribe calls. Jobber cannot list endpoints via the API, so the stored
// endpoint IDs are the only handle for later update and delete operations.
type SubscriptionResult struct {
	WebhookURL string `json:"webhookUrl"`
	// Endpoints maps object name -> topic -> created endpoint.
	Endpoints map[common.ObjectName]map[string]WebhookEndpoint `json:"endpoints"`
}

// Subscribable object names.
const (
	objectClients          = "clients"
	objectProperties       = "properties"
	objectRequests         = "requests"
	objectQuotes           = "quotes"
	objectJobs             = "jobs"
	objectVisits           = "visits"
	objectInvoices         = "invoices"
	objectExpenses         = "expenses"
	objectUsers            = "users"
	objectTimeSheetEntries = "timeSheetEntries"
	objectPayoutRecords    = "payoutRecords"
	objectProducts         = "products"
)

// objectTopicRoot maps connector object names to Jobber topic prefixes.
//
//nolint:gochecknoglobals
var objectTopicRoot = map[common.ObjectName]string{
	objectClients:          "CLIENT",
	objectProperties:       "PROPERTY",
	objectRequests:         "REQUEST",
	objectQuotes:           "QUOTE",
	objectJobs:             "JOB",
	objectVisits:           "VISIT",
	objectInvoices:         "INVOICE",
	objectExpenses:         "EXPENSE",
	objectUsers:            "USER",
	objectTimeSheetEntries: "TIMESHEET",
	objectPayoutRecords:    "PAYOUT",
	objectProducts:         "PRODUCT_OR_SERVICE",
}

// eventTypeSuffix maps framework CRUD event types to Jobber topic suffixes.
//
//nolint:gochecknoglobals
var eventTypeSuffix = map[common.SubscriptionEventType]string{
	common.SubscriptionEventTypeCreate: "_CREATE",
	common.SubscriptionEventTypeUpdate: "_UPDATE",
	common.SubscriptionEventTypeDelete: "_DESTROY",
}

// validTopics is the complete WebHookTopicEnum set, verified live via
// GraphQL introspection (X-Jobber-Graphql-Version 2025-01-20).
//
//nolint:gochecknoglobals
var validTopics = datautils.NewStringSet(
	"APP_CONNECT",
	"APP_DISCONNECT",
	"CLIENT_CREATE",
	"CLIENT_DESTROY",
	"CLIENT_UPDATE",
	"INVOICE_CREATE",
	"INVOICE_DESTROY",
	"INVOICE_UPDATE",
	"JOB_CREATE",
	"JOB_DESTROY",
	"JOB_UPDATE",
	"JOB_CLOSED",
	"PROPERTY_CREATE",
	"PROPERTY_DESTROY",
	"PROPERTY_UPDATE",
	"QUOTE_CREATE",
	"QUOTE_DESTROY",
	"QUOTE_UPDATE",
	"QUOTE_SENT",
	"QUOTE_APPROVED",
	"REQUEST_CREATE",
	"REQUEST_DESTROY",
	"REQUEST_UPDATE",
	"VISIT_COMPLETE",
	"VISIT_CREATE",
	"VISIT_DESTROY",
	"VISIT_UPDATE",
	"PRODUCT_OR_SERVICE_CREATE",
	"PRODUCT_OR_SERVICE_DESTROY",
	"PRODUCT_OR_SERVICE_UPDATE",
	"PAYMENT_CREATE",
	"PAYMENT_DESTROY",
	"PAYMENT_UPDATE",
	"PAYOUT_CREATE",
	"PAYOUT_DESTROY",
	"PAYOUT_UPDATE",
	"TIMESHEET_CREATE",
	"TIMESHEET_DESTROY",
	"TIMESHEET_UPDATE",
	"EXPENSE_CREATE",
	"EXPENSE_DESTROY",
	"EXPENSE_UPDATE",
	"ON_MY_WAY_TRACKING_LINK_REQUEST",
	"MARKETING_ITEM_UPDATE",
	"USER_CREATE",
	"USER_UPDATE",
)

// Subscribe creates one webhook endpoint per resolved topic. On partial
// failure every endpoint created so far is rolled back with a single batch
// delete.
func (c *Connector) Subscribe(
	ctx context.Context, params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	req, err := validateSubscribeRequest(c, params)
	if err != nil {
		return nil, err
	}

	pairs, err := resolveSubscriptionTopics(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	endpoints, firstErr := c.createWebhookEndpoints(ctx, pairs, req.WebhookURL)
	if firstErr != nil {
		return c.rollbackWebhookEndpoints(ctx, endpoints, firstErr)
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: params.SubscriptionEvents,
		Result: &SubscriptionResult{
			WebhookURL: req.WebhookURL,
			Endpoints:  endpoints,
		},
	}, nil
}

// UpdateSubscription reconciles the desired topic set against the previously
// stored endpoints: stale endpoints are deleted, missing ones are created,
// matching ones are kept as-is. Jobber has no endpoint update mutation.
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	req, err := validateSubscribeRequest(c, params)
	if err != nil {
		return nil, err
	}

	prev, err := extractSubscriptionResult(c, previousResult)
	if err != nil {
		return nil, err
	}

	desired, err := resolveSubscriptionTopics(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	// The webhook URL is part of each endpoint; a changed URL means every
	// existing endpoint is stale and must be recreated.
	urlChanged := req.WebhookURL != prev.WebhookURL

	kept, staleIDs, toCreate := diffEndpoints(prev.Endpoints, desired, urlChanged)

	if len(staleIDs) > 0 {
		if _, err := c.deleteWebhookEndpoints(ctx, staleIDs); err != nil {
			return nil, fmt.Errorf("failed to delete stale webhook endpoints: %w", err)
		}
	}

	created, firstErr := c.createWebhookEndpoints(ctx, toCreate, req.WebhookURL)
	if firstErr != nil {
		return c.rollbackWebhookEndpoints(ctx, created, firstErr)
	}

	final := kept
	for obj, topics := range created {
		if final[obj] == nil {
			final[obj] = make(map[string]WebhookEndpoint)
		}

		maps.Copy(final[obj], topics)
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: buildObjectEvents(final),
		Result: &SubscriptionResult{
			WebhookURL: req.WebhookURL,
			Endpoints:  final,
		},
	}, nil
}

// DeleteSubscription removes every stored webhook endpoint in one batch call.
func (c *Connector) DeleteSubscription(
	ctx context.Context, previousResult common.SubscriptionResult,
) error {
	prev, err := extractSubscriptionResult(c, &previousResult)
	if err != nil {
		return err
	}

	ids := make([]string, 0)

	for _, topics := range prev.Endpoints {
		for _, endpoint := range topics {
			ids = append(ids, endpoint.ID)
		}
	}

	if len(ids) == 0 {
		return errMissingStoredEndpoints
	}

	if _, err := c.deleteWebhookEndpoints(ctx, ids); err != nil {
		return fmt.Errorf("failed to delete webhook endpoints: %w", err)
	}

	return nil
}

// diffEndpoints splits previously created endpoints into those still wanted
// (kept) and those to remove (staleIDs), and reports which desired topics have
// no endpoint yet (toCreate). With urlChanged every endpoint is stale.
func diffEndpoints(
	previous map[common.ObjectName]map[string]WebhookEndpoint,
	desired map[string]common.ObjectName,
	urlChanged bool,
) (map[common.ObjectName]map[string]WebhookEndpoint, []string, map[string]common.ObjectName) {
	kept := make(map[common.ObjectName]map[string]WebhookEndpoint)
	staleIDs := make([]string, 0)

	for obj, topics := range previous {
		for topic, endpoint := range topics {
			if !urlChanged && desired[topic] == obj {
				if kept[obj] == nil {
					kept[obj] = make(map[string]WebhookEndpoint)
				}

				kept[obj][topic] = endpoint
			} else {
				staleIDs = append(staleIDs, endpoint.ID)
			}
		}
	}

	toCreate := make(map[string]common.ObjectName)

	for topic, obj := range desired {
		if _, exists := kept[obj][topic]; !exists {
			toCreate[topic] = obj
		}
	}

	return kept, staleIDs, toCreate
}

// createWebhookEndpoints concurrently creates one endpoint per (topic, object)
// pair. It returns every endpoint that was successfully created, together with
// the first error encountered (if any); the caller decides whether to roll back.
func (c *Connector) createWebhookEndpoints(
	ctx context.Context,
	pairs map[string]common.ObjectName,
	webhookURL string,
) (map[common.ObjectName]map[string]WebhookEndpoint, error) {
	endpoints := make(map[common.ObjectName]map[string]WebhookEndpoint)

	var (
		firstErr  error
		errorOnce sync.Once
		mutex     sync.Mutex
	)

	callbacks := make([]simultaneously.Job, 0, len(pairs))

	for topicName, objectName := range pairs {
		topic, obj := topicName, objectName

		callbacks = append(callbacks, func(ctx context.Context) error {
			endpoint, createErr := c.createWebhookEndpoint(ctx, topic, webhookURL)

			mutex.Lock()
			defer mutex.Unlock()

			if createErr != nil {
				errorOnce.Do(func() {
					firstErr = createErr
				})

				return nil
			}

			if endpoints[obj] == nil {
				endpoints[obj] = make(map[string]WebhookEndpoint)
			}

			endpoints[obj][topic] = *endpoint

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentSubscriptions, callbacks...); err != nil {
		return endpoints, fmt.Errorf("%w: %w", errFailedToCreateSubscriptions, err)
	}

	return endpoints, firstErr
}

// rollbackWebhookEndpoints deletes endpoints created during a failed
// Subscribe/UpdateSubscription attempt and reports the appropriate status.
func (c *Connector) rollbackWebhookEndpoints(
	ctx context.Context,
	endpoints map[common.ObjectName]map[string]WebhookEndpoint,
	cause error,
) (*common.SubscriptionResult, error) {
	ids := make([]string, 0)

	for _, topics := range endpoints {
		for _, endpoint := range topics {
			ids = append(ids, endpoint.ID)
		}
	}

	if len(ids) == 0 {
		return &common.SubscriptionResult{Status: common.SubscriptionStatusFailed}, cause
	}

	if _, err := c.deleteWebhookEndpoints(ctx, ids); err != nil {
		return &common.SubscriptionResult{
				Status:       common.SubscriptionStatusFailedToRollback,
				ObjectEvents: buildObjectEvents(endpoints),
				Result: &SubscriptionResult{
					Endpoints: endpoints,
				},
			}, errors.Join(
				cause,
				fmt.Errorf("failed to rollback webhook endpoints: %w", err),
			)
	}

	return &common.SubscriptionResult{Status: common.SubscriptionStatusFailed}, cause
}

type graphQLUserError struct {
	Message string   `json:"message"`
	Path    []string `json:"path"`
}

type webhookEndpointCreateResponse struct {
	Data struct {
		WebhookEndpointCreate struct {
			WebhookEndpoint *WebhookEndpoint   `json:"webhookEndpoint"`
			UserErrors      []graphQLUserError `json:"userErrors"`
		} `json:"webhookEndpointCreate"`
	} `json:"data"`
	Errors []ErrorDetails `json:"errors"`
}

type webhookEndpointDeleteResponse struct {
	Data struct {
		WebhookEndpointDelete struct {
			DeletedWebhookEndpoints []WebhookEndpoint  `json:"deletedWebhookEndpoints"`
			UserErrors              []graphQLUserError `json:"userErrors"`
		} `json:"webhookEndpointDelete"`
	} `json:"data"`
	Errors []ErrorDetails `json:"errors"`
}

func (c *Connector) createWebhookEndpoint(
	ctx context.Context, topic, webhookURL string,
) (*WebhookEndpoint, error) {
	mutation, err := graphql.Operation(queryFiles, "mutation", "webhookEndpointCreate", nil)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]any{
		gqlQueryKey: mutation,
		gqlVariablesKey: map[string]any{
			"topic": topic,
			"url":   webhookURL,
		},
	}

	resp, err := c.JSONHTTPClient().Post(ctx, c.ProviderInfo().BaseURL, requestBody, versionHeader())
	if err != nil {
		return nil, fmt.Errorf("create webhook endpoint for topic %s: %w", topic, err)
	}

	parsed, err := common.UnmarshalJSON[webhookEndpointCreateResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("parse webhook endpoint create response: %w", err)
	}

	if err := graphQLResponseError(topic, parsed.Errors, parsed.Data.WebhookEndpointCreate.UserErrors); err != nil {
		return nil, err
	}

	if parsed.Data.WebhookEndpointCreate.WebhookEndpoint == nil {
		return nil, fmt.Errorf("%w: topic %s", errEmptyWebhookEndpoint, topic)
	}

	return parsed.Data.WebhookEndpointCreate.WebhookEndpoint, nil
}

func (c *Connector) deleteWebhookEndpoints(
	ctx context.Context, endpointIDs []string,
) ([]WebhookEndpoint, error) {
	mutation, err := graphql.Operation(queryFiles, "mutation", "webhookEndpointDelete", nil)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]any{
		gqlQueryKey: mutation,
		gqlVariablesKey: map[string]any{
			"ids": endpointIDs,
		},
	}

	resp, err := c.JSONHTTPClient().Post(ctx, c.ProviderInfo().BaseURL, requestBody, versionHeader())
	if err != nil {
		return nil, fmt.Errorf("delete webhook endpoints: %w", err)
	}

	parsed, err := common.UnmarshalJSON[webhookEndpointDeleteResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("parse webhook endpoint delete response: %w", err)
	}

	deletion := parsed.Data.WebhookEndpointDelete
	if err := graphQLResponseError("webhookEndpointDelete", parsed.Errors, deletion.UserErrors); err != nil {
		return nil, err
	}

	return deletion.DeletedWebhookEndpoints, nil
}

func versionHeader() common.Header {
	return common.Header{Key: "X-Jobber-Graphql-Version", Value: apiVersion}
}

func graphQLResponseError(operation string, errs []ErrorDetails, userErrs []graphQLUserError) error {
	if len(errs) > 0 {
		messages := make([]string, len(errs))
		for i, e := range errs {
			messages[i] = e.Message
		}

		return fmt.Errorf("%w: %s: %s", common.ErrBadRequest, operation, strings.Join(messages, "; "))
	}

	if len(userErrs) > 0 {
		messages := make([]string, len(userErrs))
		for i, e := range userErrs {
			messages[i] = e.Message
		}

		return fmt.Errorf("%w: %s: %s", errSubscriptionUserErrors, operation, strings.Join(messages, "; "))
	}

	return nil
}

func validateSubscribeRequest(c *Connector, params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: Request is nil", errMissingSubscribeParams)
	}

	req, err := c.TypedSubscriptionRequest(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidSubscribeRequest, err)
	}

	if err := validator.New().Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidSubscribeRequest, err)
	}

	return req, nil
}

func extractSubscriptionResult(
	c *Connector, result *common.SubscriptionResult,
) (*SubscriptionResult, error) {
	if result == nil || result.Result == nil {
		return nil, fmt.Errorf("%w: previousResult.Result is nil", errMissingSubscribeParams)
	}

	prev, err := c.TypedSubscriptionResult(*result)
	if err != nil {
		return nil, err
	}

	if len(prev.Endpoints) == 0 {
		return nil, errMissingStoredEndpoints
	}

	return prev, nil
}

// resolveSubscriptionTopics translates the requested object events into
// Jobber webhook topics, keyed by topic with the owning object as value.
func resolveSubscriptionTopics(
	events map[common.ObjectName]common.ObjectEvents,
) (map[string]common.ObjectName, error) {
	topics := make(map[string]common.ObjectName)

	for obj, objEvents := range events {
		root, supported := objectTopicRoot[obj]
		if !supported {
			return nil, fmt.Errorf("%w: %s", errUnsupportedSubscribeObject, obj)
		}

		for _, evt := range objEvents.Events {
			suffix, ok := eventTypeSuffix[evt]
			if !ok {
				return nil, fmt.Errorf("%w: object=%s event=%s", errUnsupportedSubscribeEvent, obj, evt)
			}

			topic := root + suffix
			// Catches combinations absent from WebHookTopicEnum, e.g. USER_DESTROY.
			if !validTopics.Has(topic) {
				return nil, fmt.Errorf("%w: object=%s event=%s", errUnsupportedSubscribeEvent, obj, evt)
			}

			topics[topic] = obj
		}

		for _, raw := range objEvents.PassThroughEvents {
			if !validTopics.Has(raw) {
				return nil, fmt.Errorf("%w: %s", errUnknownPassThroughTopic, raw)
			}

			topics[raw] = obj
		}
	}

	if len(topics) == 0 {
		return nil, errNoTopicsResolved
	}

	return topics, nil
}

// buildObjectEvents converts stored endpoints back into the normalized
// ObjectEvents view. CRUD topics map to Events; every other topic (e.g.
// QUOTE_SENT, JOB_CLOSED) is reported as a PassThroughEvent.
func buildObjectEvents(
	endpoints map[common.ObjectName]map[string]WebhookEndpoint,
) map[common.ObjectName]common.ObjectEvents {
	objectEvents := make(map[common.ObjectName]common.ObjectEvents, len(endpoints))

	for obj, topics := range endpoints {
		events := common.ObjectEvents{}

		root := objectTopicRoot[obj]

		for topic := range topics {
			evt, isCrud := crudEventFromTopic(root, topic)
			if isCrud {
				events.Events = append(events.Events, evt)
			} else {
				events.PassThroughEvents = append(events.PassThroughEvents, topic)
			}
		}

		objectEvents[obj] = events
	}

	return objectEvents
}

// crudEventFromTopic reports which CRUD event a topic represents for the
// given topic root; ok is false for pass-through topics like QUOTE_SENT.
func crudEventFromTopic(root, topic string) (common.SubscriptionEventType, bool) {
	for evt, suffix := range eventTypeSuffix {
		if topic == root+suffix {
			return evt, true
		}
	}

	return common.SubscriptionEventTypeOther, false
}
