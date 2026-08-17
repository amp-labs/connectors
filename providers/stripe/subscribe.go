package stripe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/stripe/internal/metadata"
	"github.com/amp-labs/connectors/providers/stripe/internal/webhook"
	"github.com/go-playground/validator"
)

var _ connectors.SubscribeConnector = &Connector{}

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

// Subscribe creates webhook endpoint subscriptions for the specified objects and events.
// Stripe allows multiple events per endpoint, so we create one endpoint with all requested events.
// Endpoint: POST /v1/webhook_endpoints
// Doc URL: https://docs.stripe.com/api/webhook_endpoints/create
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	payload, err := buildWebhookPayloadFromParams(params)
	if err != nil {
		return nil, err
	}

	response, err := c.createWebhookEndpoint(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook endpoint: %w", err)
	}

	result, err := buildSubscriptionResult(response, params.SubscriptionEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to build subscription result: %w", err)
	}

	return result, nil
}

// buildWebhookPayloadFromParams validates the request and builds a webhook payload
// with enabled events derived from the provided subscription params.
func buildWebhookPayloadFromParams(
	params common.SubscribeParams,
) (*WebhookPayload, error) {
	req, err := validateRequest(params)
	if err != nil {
		return nil, err
	}

	requestedEventsSet, err := buildRequestedEventSet(params.SubscriptionEvents)
	if err != nil {
		return nil, err
	}

	if requestedEventsSet.IsEmpty() {
		return nil, fmt.Errorf("%w: no events to subscribe to", errMissingParams)
	}

	payload := &WebhookPayload{
		URL:           req.WebhookEndPoint,
		EnabledEvents: requestedEventsSet.List(),
	}

	return payload, nil
}

// buildRequestedEventSet builds a set of requested events from subscription events.
// Every object must resolve to at least one Stripe event; otherwise the object would be
// reported as successfully subscribed while no event is enabled for it on the endpoint.
func buildRequestedEventSet(
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
) (datautils.Set[string], error) {
	requestedEventsSet := make(datautils.Set[string])

	for obj, events := range subscriptionEvents {
		stripeEventNames, err := webhook.BuildStripeEventNames(obj, events)
		if err != nil {
			return nil, err
		}

		if len(stripeEventNames) == 0 {
			return nil, fmt.Errorf("%w: object %s has no events to subscribe to", errMissingParams, obj)
		}

		requestedEventsSet.Add(stripeEventNames)
	}

	return requestedEventsSet, nil
}

// buildSubscriptionResult builds a subscription result from the response and subscription events.
// each object gets a copy of the endpoint response, but with only its own events in EnabledEvents.
// The ID is made unique per object by concatenating endpointID:objectName.
func buildSubscriptionResult(
	response *WebhookResponse,
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
) (*common.SubscriptionResult, error) {
	subResult := &SubscriptionResult{
		WebhookId:     response.ID,
		Secret:        response.Secret,
		Subscriptions: make(map[common.ObjectName][]string),
	}

	for objectName, objectEvents := range subscriptionEvents {
		eventNames, err := webhook.BuildStripeEventNames(objectName, objectEvents)
		if err != nil {
			return nil, err
		}

		subResult.Subscriptions[objectName] = eventNames
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: subscriptionEvents,
		Result:       subResult,
	}, nil
}

func validateRequest(params common.SubscribeParams) (*SubscriptionRequest, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("%w: request is nil", errMissingParams)
	}

	req, ok := params.Request.(*SubscriptionRequest)
	if !ok {
		return nil, fmt.Errorf("%w: expected '%T' got '%T'", errInvalidRequestType, req, params.Request)
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: request is invalid: %w", errInvalidRequestType, err)
	}

	return req, nil
}

func (c *Connector) GetWebhookEndpoint(ctx context.Context, endpointID string) (*WebhookResponse, error) {
	endpointURL, err := c.getWebhookEndpointURL()
	if err != nil {
		return nil, err
	}

	endpointURL.AddPath(endpointID)

	resp, err := c.JSONHTTPClient().Get(ctx, endpointURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook endpoint: %w", err)
	}

	// Use common UnmarshalJSON utility
	result, err := common.UnmarshalJSON[WebhookResponse](resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook endpoint response: %w", err)
	}

	// Validate object field against metadata schema
	expectedObjectType, err := getExpectedObjectTypeFromMetadata("webhook_endpoints")
	if err == nil && result.Object != expectedObjectType {
		return nil, fmt.Errorf(
			"%w: expected %s from metadata, got %s",
			errObjectTypeMismatch,
			expectedObjectType,
			result.Object,
		)
	}

	return result, nil
}

// deleteWebhookEndpoint deletes a webhook endpoint by ID.
// Endpoint: DELETE /v1/webhook_endpoints/{id}
// Doc URL: https://docs.stripe.com/api/webhook_endpoints/delete
func (c *Connector) deleteWebhookEndpoint(ctx context.Context, endpointID string) error {
	url, err := c.getWebhookEndpointURL()
	if err != nil {
		return err
	}

	url.AddPath(endpointID)

	_, err = c.JSONHTTPClient().Delete(ctx, url.String())
	if err != nil {
		return err
	}

	return nil
}

func getExpectedObjectTypeFromMetadata(objectName string) (string, error) {
	objMetadata, err := metadata.Schemas.SelectOne(common.ModuleRoot, objectName)
	if err != nil {
		return "", err
	}

	fieldMetadata, ok := objMetadata.Fields["object"]
	if !ok {
		return "", fmt.Errorf("%w for %s", errObjectFieldNotFound, objectName)
	}

	if len(fieldMetadata.Values) == 0 {
		return "", fmt.Errorf("%w for %s", errNoValuesDefined, objectName)
	}

	return fieldMetadata.Values[0].Value, nil
}

// parseWebhookEndpointResponse parses and validates the webhook endpoint response.
func parseWebhookEndpointResponse(
	ctx context.Context, httpResp *http.Response, bodyBytes []byte,
) (*WebhookResponse, error) {
	// Use common JSON parsing utilities
	jsonResp, err := common.ParseJSONResponse(ctx, httpResp, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	result, err := common.UnmarshalJSON[WebhookResponse](jsonResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook endpoint response: %w", err)
	}

	expectedObjectType, err := getExpectedObjectTypeFromMetadata("webhook_endpoints")
	if err == nil && result.Object != expectedObjectType {
		return nil, fmt.Errorf(
			"%w: expected %s from metadata, got %s",
			errObjectTypeMismatch,
			expectedObjectType,
			result.Object,
		)
	}

	return result, nil
}

func (c *Connector) createWebhookEndpoint(
	ctx context.Context,
	payload *WebhookPayload,
) (*WebhookResponse, error) {
	endpointURL, err := c.getWebhookEndpointURL()
	if err != nil {
		return nil, err
	}

	formData := buildFormData(payload)

	httpResp, bodyBytes, err := c.executeFormPostRequest(ctx, endpointURL, formData, "create")
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	return parseWebhookEndpointResponse(ctx, httpResp, bodyBytes)
}

func (c *Connector) getWebhookEndpointURL() (*urlbuilder.URL, error) {
	return c.GetURL("webhook_endpoints")
}

func (c *Connector) updateWebhookEndpoint(
	ctx context.Context,
	endpointID string,
	payload *WebhookPayload,
) (*WebhookResponse, error) {
	endpointURL, err := c.getWebhookEndpointURL()
	if err != nil {
		return nil, err
	}

	endpointURL.AddPath(endpointID)

	formData := buildFormData(payload)

	httpResp, bodyBytes, err := c.executeFormPostRequest(ctx, endpointURL, formData, "update")
	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()

	return parseWebhookEndpointResponse(ctx, httpResp, bodyBytes)
}

// executeFormPostRequest executes a POST request with form-encoded data using HTTPClient.Post.
func (c *Connector) executeFormPostRequest(
	ctx context.Context,
	endpointURL *urlbuilder.URL,
	formData url.Values,
	operation string,
) (*http.Response, []byte, error) {
	formEncoded := formData.Encode()
	formBytes := []byte(formEncoded)

	httpResp, bodyBytes, err := c.HTTPClient().Post(ctx, endpointURL.String(), formBytes, common.Header{
		Key:   "Content-Type",
		Value: "application/x-www-form-urlencoded",
		Mode:  common.HeaderModeOverwrite,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to %s webhook endpoint: %w", operation, err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, nil, common.InterpretError(httpResp, bodyBytes)
	}

	return httpResp, bodyBytes, nil
}

func buildFormData(payload *WebhookPayload) url.Values {
	formData := url.Values{}
	formData.Set("url", payload.URL)

	for _, event := range payload.EnabledEvents {
		formData.Add("enabled_events[]", event)
	}

	return formData
}
