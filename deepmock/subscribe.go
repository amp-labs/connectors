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

type RegistrationParams struct {
	Metadata map[string]any `json:"metadata"`
}

type RegistrationResult struct {
	Metadata map[string]any `json:"metadata"`
}

func (c *Connector) Register(_ context.Context, params common.SubscriptionRegistrationParams) (*common.RegistrationResult, error) {
	if params.Request == nil {
		return nil, fmt.Errorf("params.Request can't be nil")
	}

	req, ok := params.Request.(*RegistrationParams)
	if !ok {
		return nil, fmt.Errorf("params.Request can't be cast to *RegistrationParams")
	}

	result := &RegistrationResult{
		Metadata: req.Metadata,
	}

	return &common.RegistrationResult{
		RegistrationRef: uuid.New().String(),
		Result:          result,
		Status:          common.RegistrationStatusSuccess,
	}, nil
}

func (c *Connector) DeleteRegistration(_ context.Context, previousResult common.RegistrationResult) error {
	if previousResult.Result == nil {
		return fmt.Errorf("previousResult.Result can't be nil")
	}

	_, ok := previousResult.Result.(*RegistrationResult)
	if !ok {
		return fmt.Errorf("previousResult.Result can't be cast to *RegistrationResult")
	}

	return nil
}

func (c *Connector) EmptyRegistrationParams() *common.SubscriptionRegistrationParams {
	return &common.SubscriptionRegistrationParams{
		Request: &RegistrationParams{},
	}
}

func (c *Connector) EmptyRegistrationResult() *common.RegistrationResult {
	return &common.RegistrationResult{
		Result: &RegistrationResult{},
	}
}

type SubscribeParams struct {
	Observer func(action string, record map[string]any) `json:"-"`
	Metadata map[string]any                             `json:"metadata"`
}

type SubscribeResult struct {
	Metadata map[string]any `json:"metadata"`
}

// Subscribe creates a new subscription for the deepmock connector.
// Since deepmock is an in-memory test connector, this is a no-op that returns success.
// Actual event observation is handled by the observer option passed during connector creation.
func (c *Connector) Subscribe(
	ctx context.Context,
	params common.SubscribeParams,
) (*common.SubscriptionResult, error) {
	if params.RegistrationResult == nil {
		return nil, fmt.Errorf("params.RegistrationResult can't be nil")
	}

	if params.RegistrationResult.Status != common.RegistrationStatusSuccess {
		return nil, fmt.Errorf("params.RegistrationResult.Status can't be %s", params.RegistrationResult.Status)
	}

	if params.RegistrationResult.Result == nil {
		return nil, fmt.Errorf("params.RegistrationResult.Result can't be nil")
	}

	regResult, ok := params.RegistrationResult.Result.(*RegistrationResult)
	if !ok {
		return nil, fmt.Errorf("params.RegistrationResult.Result can't be cast to *SubscribeResult")
	}

	if len(params.SubscriptionEvents) == 0 {
		return nil, fmt.Errorf("params.SubscriptionEvents can't be empty")
	}

	if params.Request == nil {
		return nil, fmt.Errorf("params.Request can't be nil")
	}

	request, ok := params.Request.(*SubscribeParams)
	if !ok {
		return nil, fmt.Errorf("params.Request can't be cast to *SubscribeParams")
	}

	tracker := &SubscriptionInfo{
		Id:                 uuid.New().String(),
		RegistrationRef:    params.RegistrationResult.RegistrationRef,
		RegistrationResult: regResult,
		Observer:           request.Observer,
		Metadata:           request.Metadata,
	}

	// For deepmock, we don't need to create any real subscription with a provider
	// The observer (if configured) will handle event publishing
	return &common.SubscriptionResult{
		Result: &SubscribeResult{
			Metadata: tracker.Metadata,
		},
	}, nil
}

// UpdateSubscription updates an existing subscription.
// Since deepmock subscriptions are no-ops, this also returns success.
func (c *Connector) UpdateSubscription(
	_ context.Context,
	_ common.SubscribeParams,
	_ *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	return &common.SubscriptionResult{}, nil
}

// DeleteSubscription removes a subscription.
// Since deepmock subscriptions are no-ops, this returns success.
func (c *Connector) DeleteSubscription(
	_ context.Context,
	_ common.SubscriptionResult,
) error {
	return nil
}

// EmptySubscriptionParams returns an empty SubscribeParams instance.
func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return &common.SubscribeParams{}
}

// EmptySubscriptionResult returns an empty SubscriptionResult instance.
func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return &common.SubscriptionResult{}
}

// VerifyWebhookMessage verifies a webhook message signature.
// For deepmock, we always return true since this is a test connector.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	// Always verify successfully for deepmock - this is a test connector
	return true, nil
}

// GetRecordsByIds reads multiple records by their IDs.
// This implements the BatchRecordReaderConnector interface required by WebhookVerifierConnector.
//
//nolint:revive // recordIds parameter name matches interface definition
func (c *Connector) GetRecordsByIds(
	_ context.Context,
	objectName string,
	recordIds []string,
	_ []string, // fields parameter (unused for deepmock - returns all fields)
	_ []string, // associations parameter (unused for deepmock)
) ([]common.ReadResultRow, error) {
	// Read records from storage by their IDs
	var results []common.ReadResultRow

	for _, id := range recordIds {
		record, err := c.storage.Get(objectName, id)
		if err != nil {
			if errors.Is(err, ErrRecordNotFound) {
				continue
			}

			// If record not found, skip it (consistent with provider behavior)
			return nil, err
		}

		// Convert to ReadResultRow
		row := common.ReadResultRow{
			Id:     id,
			Fields: record,
			Raw:    record,
		}

		results = append(results, row)
	}

	return results, nil
}
