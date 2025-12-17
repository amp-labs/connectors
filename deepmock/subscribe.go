package deepmock

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/google/uuid"
)

// Compile-time interface check for SubscribeConnector.
var _ connectors.SubscribeConnector = (*Connector)(nil)

// RegistrationParams contains the parameters for registering a webhook endpoint.
// For deepmock, this is a simple container for metadata that can be used to
// track registration-specific information.
type RegistrationParams struct {
	Metadata map[string]any `json:"metadata"`
}

// RegistrationResult contains the result of a successful webhook registration.
// It includes any metadata from the registration request.
type RegistrationResult struct {
	Metadata map[string]any `json:"metadata"`
}

// Register creates a registration for webhook subscriptions.
// This is the first step in the subscription flow, establishing a reference
// that can be used by subsequent Subscribe calls.
//
// For deepmock, this is a lightweight operation that generates a unique
// registration reference and returns success. The metadata from the request
// is preserved in the result for tracking purposes.
//
// Returns an error if the request is nil or has an invalid type.
func (c *Connector) Register(
	_ context.Context, params common.SubscriptionRegistrationParams,
) (*common.RegistrationResult, error) {
	if params.Request == nil {
		return nil, ErrRequestNil
	}

	req, ok := params.Request.(*RegistrationParams)
	if !ok {
		return nil, ErrInvalidRequestType
	}

	// Preserve metadata from the request
	result := &RegistrationResult{
		Metadata: req.Metadata,
	}

	return &common.RegistrationResult{
		RegistrationRef: uuid.New().String(),
		Result:          result,
		Status:          common.RegistrationStatusSuccess,
	}, nil
}

// DeleteRegistration removes a webhook registration.
// This is typically called when a subscription is no longer needed and should
// clean up any resources associated with the registration.
//
// For deepmock, this validates that the registration result has the correct
// structure but performs no actual cleanup since registrations are stateless.
//
// Returns an error if the result is nil or has an invalid type.
func (c *Connector) DeleteRegistration(_ context.Context, previousResult common.RegistrationResult) error {
	if previousResult.Result == nil {
		return ErrResultNil
	}

	_, ok := previousResult.Result.(*RegistrationResult)
	if !ok {
		return ErrInvalidResultType
	}

	return nil
}

// EmptyRegistrationParams returns an empty RegistrationParams instance.
// This is used by the connector framework for JSON unmarshaling and type checking.
func (c *Connector) EmptyRegistrationParams() *common.SubscriptionRegistrationParams {
	return &common.SubscriptionRegistrationParams{
		Request: &RegistrationParams{},
	}
}

// EmptyRegistrationResult returns an empty RegistrationResult instance.
// This is used by the connector framework for JSON unmarshaling and type checking.
func (c *Connector) EmptyRegistrationResult() *common.RegistrationResult {
	return &common.RegistrationResult{
		Result: &RegistrationResult{},
	}
}

// SubscribeParams contains the parameters for creating a subscription.
// It includes the callback function that will be invoked when subscribed events occur,
// and optional metadata for tracking subscription-specific information.
type SubscribeParams struct {
	Notify   NotifyCallback `json:"-"`        // Callback invoked for matching events (excluded from JSON)
	Metadata map[string]any `json:"metadata"` // Optional metadata for tracking
}

// SubscribeResult contains the result of a successful subscription.
// It includes the full subscription context with all configuration details.
type SubscribeResult struct {
	Subscription *SubscriptionContext `json:"subscription"`
}

// Subscribe creates a new subscription for the deepmock connector.
// This registers the subscription with the storage layer so that the notify
// callback will be invoked when matching storage operations occur.
//
// The subscription filters events based on:
//   - Object types (e.g., "contact", "account")
//   - Event types (create, update, delete)
//   - Specific fields that changed (for update events)
//
// Notifications are delivered asynchronously in separate goroutines to prevent
// blocking storage operations.
//
// Returns an error if:
//   - The registration result is nil or invalid
//   - No subscription events are specified
//   - The request parameters are nil or invalid
//   - Storage subscription fails
func (c *Connector) Subscribe(
	_ context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if params.RegistrationResult == nil {
		return nil, ErrRegistrationResultNil
	}

	if params.RegistrationResult.Status != common.RegistrationStatusSuccess {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRegistrationStatus, params.RegistrationResult.Status)
	}

	if params.RegistrationResult.Result == nil {
		return nil, ErrResultNil
	}

	// Extract and validate the registration result
	regResult, isValid := params.RegistrationResult.Result.(*RegistrationResult)
	if !isValid {
		return nil, ErrInvalidResultType
	}

	if len(params.SubscriptionEvents) == 0 {
		return nil, ErrSubscriptionEventsEmpty
	}

	if params.Request == nil {
		return nil, ErrRequestNil
	}

	// Extract subscription parameters
	request, ok := params.Request.(*SubscribeParams)
	if !ok {
		return nil, ErrInvalidRequestType
	}

	// Create subscription context with a unique ID
	sub := &SubscriptionContext{
		Notify:             request.Notify,
		Id:                 uuid.New().String(),
		SubscriptionEvents: params.SubscriptionEvents,
		RegistrationRef:    params.RegistrationResult.RegistrationRef,
		RegistrationResult: regResult,
		Metadata:           request.Metadata,
	}

	// Register the subscription with storage to start receiving notifications
	if err := c.storage.Subscribe(sub); err != nil {
		return nil, err
	}

	// Generate summary information for the subscription result
	objects, events, updateFields := c.getSubscribeSummary(params)

	return &common.SubscriptionResult{
		Result: &SubscribeResult{
			Subscription: sub,
		},
		ObjectEvents: params.SubscriptionEvents,
		Status:       common.SubscriptionStatusSuccess,
		Objects:      objects,
		Events:       events,
		UpdateFields: updateFields,
	}, nil
}

// getSubscribeSummary extracts summary information from subscription parameters.
// This helper function analyzes the subscription events to determine:
//  1. Which objects are being subscribed to
//  2. Which event types are being watched (deduplicated across all objects)
//  3. Which fields are available for those objects (for field-level filtering)
//
// The returned information is included in the SubscriptionResult to inform
// the caller about what the subscription covers.
//
//nolint:cyclop,funlen // Complexity from processing multiple subscription event types
func (c *Connector) getSubscribeSummary(
	params common.SubscribeParams,
) ([]common.ObjectName, []common.SubscriptionEventType, []string) {
	var objects []common.ObjectName

	// Extract all object names from the subscription events
	if len(params.SubscriptionEvents) > 0 {
		objects = make([]common.ObjectName, 0, len(params.SubscriptionEvents))

		for objName := range params.SubscriptionEvents {
			objects = append(objects, objName)
		}
	}

	// Deduplicate event types across all objects
	eventsDedup := make(map[common.SubscriptionEventType]struct{})

	for _, evts := range params.SubscriptionEvents {
		for _, evt := range evts.Events {
			eventsDedup[evt] = struct{}{}
		}
	}

	var events []common.SubscriptionEventType

	if len(eventsDedup) > 0 {
		events = make([]common.SubscriptionEventType, 0, len(eventsDedup))

		for evt := range eventsDedup {
			events = append(events, evt)
		}
	}

	// Collect all available fields from the schemas of subscribed objects
	updateFieldsDedup := make(map[string]struct{})

	for _, objectName := range objects {
		schema, exists := c.schemas.Get(string(objectName))
		if !exists {
			continue
		}

		// Get associations for this object (if any)
		associations := c.storage.GetAssociations()[ObjectName(objectName)]

		metadata := schemaToObjectMetadata(string(objectName), schema, associations)
		if metadata == nil {
			continue
		}

		// Add all fields from this object's schema
		for field := range metadata.Fields {
			updateFieldsDedup[field] = struct{}{}
		}
	}

	var updateFields []string

	if len(updateFieldsDedup) > 0 {
		updateFields = make([]string, 0, len(updateFieldsDedup))

		for field := range updateFieldsDedup {
			updateFields = append(updateFields, field)
		}
	}

	return objects, events, updateFields
}

// UpdateSubscription updates an existing subscription by removing the old one and creating a new one.
// This allows changing the subscription's events, fields, or callback without losing the registration.
func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	// First, delete the old subscription
	//nolint:nestif // Complex nested logic for subscription update
	if previousResult != nil && previousResult.Result != nil {
		prevResult, ok := previousResult.Result.(*SubscribeResult)
		if ok && prevResult.Subscription != nil {
			// Remove the old subscription from storage
			if err := c.storage.Unsubscribe(prevResult.Subscription.Id); err != nil {
				// If subscription doesn't exist, that's fine - we're updating anyway
				if !errors.Is(err, ErrObserverNotFound) {
					return nil, fmt.Errorf("failed to unsubscribe previous subscription: %w", err)
				}
			}
		}
	}

	// Now create the new subscription using the same logic as Subscribe
	return c.Subscribe(ctx, params)
}

// DeleteSubscription removes a subscription from storage, stopping all future notifications.
func (c *Connector) DeleteSubscription(
	ctx context.Context,
	result common.SubscriptionResult,
) error {
	// Extract the subscription ID from the result
	if result.Result == nil {
		return ErrResultNil
	}

	subResult, ok := result.Result.(*SubscribeResult)
	if !ok {
		return ErrInvalidResultType
	}

	if subResult.Subscription == nil {
		return ErrSubscriptionNil
	}

	// Unsubscribe from storage
	if err := c.storage.Unsubscribe(subResult.Subscription.Id); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	return nil
}

// EmptySubscriptionParams returns an empty SubscribeParams instance.
// This is used by the connector framework for JSON unmarshaling and type checking.
// Note: The Request field contains SubscribeParams (not RegistrationParams) since
// this is for the subscription phase, not the registration phase.
func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{
		Request: &SubscribeParams{},
	}
}

// EmptySubscriptionResult returns an empty SubscriptionResult instance.
// This is used by the connector framework for JSON unmarshaling and type checking.
// Note: The Result field contains SubscribeResult (not RegistrationResult) since
// this represents the result of a subscription, not a registration.
func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{
		Result: &SubscribeResult{},
	}
}

// VerifyWebhookMessage verifies a webhook message signature.
// For deepmock, this always returns true since it's a test connector without
// real webhook security requirements.
//
// In a real connector implementation, this would validate:
//   - HMAC signatures or other cryptographic verification
//   - Timestamp freshness to prevent replay attacks
//   - Request origin and headers
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	// Always verify successfully for deepmock - this is a test connector
	return true, nil
}

// GetRecordsByIds reads multiple records by their IDs from storage.
// This implements the BatchRecordReaderConnector interface required by WebhookVerifierConnector,
// enabling webhook verification workflows that need to fetch current record state.
//
// The method retrieves records from the in-memory storage and returns them in a format
// compatible with the connector framework. Records that don't exist are silently skipped
// (consistent with real provider behavior).
//
// Parameters:
//   - objectName: The type of object to retrieve (e.g., "contact", "account")
//   - recordIds: List of record IDs to fetch
//   - fields: List of field names to include in the response; if empty/nil, all fields are returned
//   - associations: List of association names to expand (e.g., "account", "contacts")
//
// Returns a slice of ReadResultRow containing the found records. The Fields map contains
// only the requested fields (or all fields if none specified), while Raw always contains
// the complete record. If associations are requested, they will be populated in the
// Associations field of each row. Returns an error if the storage operation fails.
//
//nolint:revive,cyclop,nestif,funlen // recordIds parameter name; complexity from filtering
func (c *Connector) GetRecordsByIds(
	_ context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	// First, collect all the records
	records := make([]map[string]any, 0, len(recordIds))
	results := make([]common.ReadResultRow, 0, len(recordIds))

	for _, id := range recordIds {
		record, err := c.storage.Get(objectName, id)
		if err != nil {
			// Skip records that don't exist (consistent with provider behavior)
			if errors.Is(err, ErrRecordNotFound) {
				continue
			}

			// Other errors should be propagated
			return nil, err
		}

		records = append(records, record)
	}

	// Expand associations if requested
	var associationsMap map[string]map[string][]common.Association

	if len(associations) > 0 && len(records) > 0 {
		var err error

		associationsMap, err = c.expandAssociations(objectName, records, associations)
		if err != nil {
			return nil, fmt.Errorf("failed to expand associations: %w", err)
		}
	}

	// Get the ID field name for extracting record IDs
	idField := c.storage.GetIdFields()[ObjectName(objectName)]
	if idField == "" {
		idField = "id"
	}

	// Build result rows
	for _, record := range records {
		// Get record ID for association lookup
		recordID, ok := record[idField]
		if !ok {
			continue
		}

		recordIDStr := fmt.Sprintf("%v", recordID)

		// Filter fields if requested
		filteredFields := record
		if len(fields) > 0 {
			filteredFields = make(map[string]any, len(fields))
			for _, field := range fields {
				if value, exists := record[field]; exists {
					filteredFields[field] = value
				}
			}
		}

		// Convert to ReadResultRow format
		// Fields contains only requested fields (or all if none specified)
		// Raw always contains the complete record
		row := common.ReadResultRow{
			Id:     recordIDStr,
			Fields: filteredFields,
			Raw:    record,
		}

		// Add associations if present
		if associationsMap != nil {
			if recordAssocs, exists := associationsMap[recordIDStr]; exists {
				row.Associations = recordAssocs
			}
		}

		results = append(results, row)
	}

	return results, nil
}
