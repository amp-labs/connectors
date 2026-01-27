package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
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
		Result: &SubscriptionResult{
			Subscriptions: make(map[common.ObjectName]WebhookResponse),
		},
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

	if len(requestedEventsSet) == 0 {
		return nil, fmt.Errorf("%w: no events to subscribe to", errMissingParams)
	}

	enabledEvents := make([]string, 0, len(requestedEventsSet))
	for event := range requestedEventsSet {
		enabledEvents = append(enabledEvents, event)
	}

	payload := &WebhookPayload{
		URL:           req.WebhookEndPoint,
		EnabledEvents: enabledEvents,
	}

	return payload, nil
}

// buildRequestedEventSet builds a set of requested events from subscription events.
func buildRequestedEventSet(subscriptionEvents map[common.ObjectName]common.ObjectEvents) (map[string]bool, error) {
	requestedEventsSet := make(map[string]bool)

	for obj, events := range subscriptionEvents {
		for _, event := range events.Events {
			stripeEventName, err := getStripeEventName(event, obj)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event type %s for object %s: %w", event, obj, err)
			}

			requestedEventsSet[stripeEventName] = true
		}

		// Add pass-through events directly
		for _, passthroughEvent := range events.PassThroughEvents {
			requestedEventsSet[passthroughEvent] = true
		}
	}

	return requestedEventsSet, nil
}

// extractBaseEndpointID extracts the base endpoint ID from a composite ID.
// Composite IDs are in the format "endpointID:objectName", so we extract everything before the last colon.
// If no colon is found, the ID is returned as-is (for backward compatibility).
func extractBaseEndpointID(compositeID string) string {
	if idx := strings.LastIndex(compositeID, ":"); idx != -1 {
		return compositeID[:idx]
	}

	return compositeID
}

// buildSubscriptionResult builds a subscription result from the response and subscription events.
// each object gets a copy of the endpoint response, but with only its own events in EnabledEvents.
// The ID is made unique per object by concatenating endpointID:objectName.
func buildSubscriptionResult(
	response *WebhookResponse,
	subscriptionEvents map[common.ObjectName]common.ObjectEvents,
) (*common.SubscriptionResult, error) {
	subscriptionsMap := make(map[common.ObjectName]WebhookResponse)

	for obj, events := range subscriptionEvents {
		// Filter enabled events to only include events for this object
		objectEvents := make([]string, 0)

		for _, event := range events.Events {
			stripeEventName, err := getStripeEventName(event, obj)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event type %s for object %s: %w", event, obj, err)
			}

			objectEvents = append(objectEvents, stripeEventName)
		}

		// Add pass-through events directly
		objectEvents = append(objectEvents, events.PassThroughEvents...)

		objectResponse := *response
		objectResponse.EnabledEvents = objectEvents
		// Make ID unique per object: endpointID:objectName
		objectResponse.ID = fmt.Sprintf("%s:%s", response.ID, string(obj))
		subscriptionsMap[obj] = objectResponse
	}

	return &common.SubscriptionResult{
		Status:       common.SubscriptionStatusSuccess,
		ObjectEvents: subscriptionEvents,
		Result: &SubscriptionResult{
			Subscriptions: subscriptionsMap,
		},
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

	resp, err := c.Client.Get(ctx, endpointURL.String())
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

	_, err = c.Client.Delete(ctx, url.String())
	if err != nil {
		return err
	}

	return nil
}

// getStripeEventName converts normalized CRUD events to Stripe event type strings.
// It only generates Stripe's standard *.created / *.updated / *.deleted actions.
// For any other actions, callers must provide full Stripe event names via PassThroughEvents.
// Doc URL: https://docs.stripe.com/api/events/types
func getStripeEventName(event common.SubscriptionEventType, obj common.ObjectName) (string, error) {
	objectName := strings.ToLower(string(obj))

	switch event {
	case common.SubscriptionEventTypeCreate:
		return objectName + ".created", nil
	case common.SubscriptionEventTypeUpdate:
		return objectName + ".updated", nil
	case common.SubscriptionEventTypeDelete:
		return objectName + ".deleted", nil
	case common.SubscriptionEventTypeAssociationUpdate,
		common.SubscriptionEventTypeOther:
		return "", fmt.Errorf("%w: %s", errUnsupportedEventType, event)
	default:
		return "", fmt.Errorf("%w: %s", errUnsupportedEventType, event)
	}
}

func getExpectedObjectTypeFromMetadata(objectName string) (string, error) {
	objMetadata, err := metadata.Schemas.SelectOne(common.ModuleID("root"), objectName)
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
func parseWebhookEndpointResponse(httpResp *http.Response, bodyBytes []byte) (*WebhookResponse, error) {
	// Use common JSON parsing utilities
	jsonResp, err := common.ParseJSONResponse(httpResp, bodyBytes)
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

	return parseWebhookEndpointResponse(httpResp, bodyBytes)
}

func (c *Connector) getWebhookEndpointURL() (*urlbuilder.URL, error) {
	return urlbuilder.New(c.BaseURL, apiVersion, "webhook_endpoints")
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

	return parseWebhookEndpointResponse(httpResp, bodyBytes)
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

	httpResp, bodyBytes, err := c.Client.HTTPClient.Post(ctx, endpointURL.String(), formBytes, common.Header{
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

const (
	stripeSignatureHeader = "Stripe-Signature"
	keyValuePairParts     = 2
	defaultTolerance      = 5 * time.Minute
)

// VerifyWebhookMessage verifies the signature of a webhook message from Stripe.
// Stripe uses HMAC-SHA256 with the format: signed_payload = timestamp + "." + requestBody.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	if request == nil || params == nil {
		return false, fmt.Errorf("%w: request and params cannot be nil", errMissingParams)
	}

	verificationParams, err := common.AssertType[*VerificationParams](params.Param)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errMissingParams, err)
	}

	if verificationParams.Secret == "" {
		return false, fmt.Errorf("%w: secret cannot be empty", errMissingParams)
	}

	// Determine tolerance: use provided value or default to 5 minutes
	tolerance := verificationParams.Tolerance
	if tolerance == 0 {
		tolerance = defaultTolerance
	} else if tolerance < 0 {
		return false, fmt.Errorf("%w", errInvalidTolerance)
	}

	signatureHeader := request.Headers.Get(stripeSignatureHeader)
	if signatureHeader == "" {
		return false, fmt.Errorf("%w: missing %s header", errMissingSignature, stripeSignatureHeader)
	}

	timestampStr, signatures, err := parseStripeSignature(signatureHeader)
	if err != nil {
		return false, fmt.Errorf("failed to parse Stripe signature: %w", err)
	}

	if err := validateTimestampRecency(timestampStr, tolerance); err != nil {
		return false, err
	}

	return verifySignature(verificationParams.Secret, timestampStr, request.Body, signatures)
}

// verifySignature computes the expected HMAC-SHA256 signature and compares it with the received signatures.
func verifySignature(secret, timestampStr string, body []byte, receivedSignatures []string) (bool, error) {
	// Create signed_payload: timestamp + "." + requestBody
	signedPayload := fmt.Sprintf("%s.%s", timestampStr, string(body))

	// Compute expected signature using HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant-time comparison
	// Check if any of the received signatures matches the expected signature
	for _, sig := range receivedSignatures {
		if hmac.Equal([]byte(expectedSignature), []byte(sig)) {
			return true, nil
		}
	}

	return false, fmt.Errorf("%w: signature mismatch", errInvalidSignature)
}

// validateTimestampRecency validates that the webhook timestamp is within the allowed tolerance.
func validateTimestampRecency(timestampStr string, tolerance time.Duration) error {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	timestampTime := time.Unix(timestamp, 0)
	now := time.Now()
	timeDiff := now.Sub(timestampTime)

	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	if timeDiff > tolerance {
		if timestampTime.Before(now) {
			return errTimestampTooOld
		}

		return errTimestampTooFarInFuture
	}

	return nil
}

func parseStripeSignature(header string) (string, []string, error) {
	elements := strings.Split(header, ",")

	var timestamp string

	var signatures []string

	for _, element := range elements {
		element = strings.TrimSpace(element)

		parts := strings.SplitN(element, "=", keyValuePairParts)
		if len(parts) != keyValuePairParts {
			continue
		}

		prefix := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch prefix {
		case "t":
			timestamp = value
		case "v1", "v2", "v0":
			signatures = append(signatures, value)
		}
	}

	if timestamp == "" {
		return "", nil, fmt.Errorf("%w", errMissingTimestamp)
	}

	if len(signatures) == 0 {
		return "", nil, fmt.Errorf("%w", errNoSignaturesFound)
	}

	return timestamp, signatures, nil
}
