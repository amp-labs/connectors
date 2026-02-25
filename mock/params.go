package mock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// Option is a function which mutates the hubspot connector configuration.
type Option = func(params *parameters)

// WithClient sets the http client to use for the connector. Saves some boilerplate.
func WithClient(client *http.Client) Option {
	return func(params *parameters) {
		WithAuthenticatedClient(client)(params)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: client,
			},
		}
	}
}

// WithRead sets the read function for the connector.
func WithRead(read func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)) Option {
	return func(params *parameters) {
		params.read = read
	}
}

// WithWrite sets the write function for the connector.
func WithWrite(write func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)) Option {
	return func(params *parameters) {
		params.write = write
	}
}

// WithListObjectMetadata sets the listObjectMetadata function for the connector.
func WithListObjectMetadata(
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error),
) Option {
	return func(params *parameters) {
		params.listObjectMetadata = listObjectMetadata
	}
}

// WithGetURL sets the getURL function for the connector.
func WithGetURL(
	getURL func(resource string, args map[string]any) (string, error),
) Option {
	return func(params *parameters) {
		params.getURL = getURL
	}
}

// WithDelete sets the delete function for the connector.
func WithDelete(
	deleteFunc func(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error),
) Option {
	return func(params *parameters) {
		params.delete = deleteFunc
	}
}

// WithGetPostAuthInfo sets the getPostAuthInfo function for the connector.
func WithGetPostAuthInfo(
	getPostAuthInfo func(ctx context.Context) (*common.PostAuthInfo, error),
) Option {
	return func(params *parameters) {
		params.getPostAuthInfo = getPostAuthInfo
	}
}

// WithGetRecordsByIds sets the getRecordsByIds function for the connector.
func WithGetRecordsByIds(
	getRecordsByIds func(ctx context.Context, params common.ReadByIdsParams) ([]common.ReadResultRow, error),
) Option {
	return func(params *parameters) {
		params.getRecordsByIds = getRecordsByIds
	}
}

// WithVerifyWebhookMessage sets the verifyWebhookMessage function for the connector.
func WithVerifyWebhookMessage(
	verifyWebhookMessage func(
		ctx context.Context,
		request *common.WebhookRequest,
		params *common.VerificationParams,
	) (bool, error),
) Option {
	return func(params *parameters) {
		params.verifyWebhookMessage = verifyWebhookMessage
	}
}

// WithRegister sets the register function for the connector.
func WithRegister(
	register func(ctx context.Context, params common.SubscriptionRegistrationParams) (*common.RegistrationResult, error),
) Option {
	return func(params *parameters) {
		params.register = register
	}
}

// WithDeleteRegistration sets the deleteRegistration function for the connector.
func WithDeleteRegistration(
	deleteRegistration func(ctx context.Context, previousResult common.RegistrationResult) error,
) Option {
	return func(params *parameters) {
		params.deleteRegistration = deleteRegistration
	}
}

// WithEmptyRegistrationParams sets the emptyRegistrationParams function for the connector.
func WithEmptyRegistrationParams(
	emptyRegistrationParams func() *common.SubscriptionRegistrationParams,
) Option {
	return func(params *parameters) {
		params.emptyRegistrationParams = emptyRegistrationParams
	}
}

// WithEmptyRegistrationResult sets the emptyRegistrationResult function for the connector.
func WithEmptyRegistrationResult(
	emptyRegistrationResult func() *common.RegistrationResult,
) Option {
	return func(params *parameters) {
		params.emptyRegistrationResult = emptyRegistrationResult
	}
}

// WithSubscribe sets the subscribe function for the connector.
func WithSubscribe(
	subscribe func(ctx context.Context, params common.SubscribeParams) (*common.SubscriptionResult, error),
) Option {
	return func(params *parameters) {
		params.subscribe = subscribe
	}
}

// WithUpdateSubscription sets the updateSubscription function for the connector.
func WithUpdateSubscription(updateSubscription func(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error),
) Option {
	return func(params *parameters) {
		params.updateSubscription = updateSubscription
	}
}

// WithDeleteSubscription sets the deleteSubscription function for the connector.
func WithDeleteSubscription(
	deleteSubscription func(ctx context.Context, previousResult common.SubscriptionResult) error,
) Option {
	return func(params *parameters) {
		params.deleteSubscription = deleteSubscription
	}
}

// WithEmptySubscriptionParams sets the emptySubscriptionParams function for the connector.
func WithEmptySubscriptionParams(
	emptySubscriptionParams func() *common.SubscribeParams,
) Option {
	return func(params *parameters) {
		params.emptySubscriptionParams = emptySubscriptionParams
	}
}

// WithEmptySubscriptionResult sets the emptySubscriptionResult function for the connector.
func WithEmptySubscriptionResult(
	emptySubscriptionResult func() *common.SubscriptionResult,
) Option {
	return func(params *parameters) {
		params.emptySubscriptionResult = emptySubscriptionResult
	}
}

// parameters is the internal configuration for the mock connector.
type parameters struct {
	client             *common.JSONHTTPClient // required
	read               func(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
	write              func(ctx context.Context, params common.WriteParams) (*common.WriteResult, error)
	listObjectMetadata func(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error)
	getURL             func(resource string, args map[string]any) (string, error)
	delete             func(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error)
	getPostAuthInfo    func(ctx context.Context) (*common.PostAuthInfo, error)

	getRecordsByIds func(ctx context.Context, params common.ReadByIdsParams) ([]common.ReadResultRow, error)

	verifyWebhookMessage func(
		ctx context.Context,
		request *common.WebhookRequest,
		params *common.VerificationParams,
	) (bool, error)

	register func(
		ctx context.Context,
		params common.SubscriptionRegistrationParams,
	) (*common.RegistrationResult, error)

	deleteRegistration      func(ctx context.Context, previousResult common.RegistrationResult) error
	emptyRegistrationParams func() *common.SubscriptionRegistrationParams
	emptyRegistrationResult func() *common.RegistrationResult
	subscribe               func(ctx context.Context, params common.SubscribeParams) (*common.SubscriptionResult, error)

	updateSubscription func(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)

	deleteSubscription      func(ctx context.Context, previousResult common.SubscriptionResult) error
	emptySubscriptionParams func() *common.SubscribeParams
	emptySubscriptionResult func() *common.SubscriptionResult
}

func (p parameters) ValidateParams() error { //nolint:funlen,cyclop
	if p.client == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "client")
	}

	if p.read == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "read")
	}

	if p.write == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "write")
	}

	if p.listObjectMetadata == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "listObjectMetadata")
	}

	if p.getURL == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "getURL")
	}

	if p.delete == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "delete")
	}

	if p.getPostAuthInfo == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "getPostAuthInfo")
	}

	if p.getRecordsByIds == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "getRecordsByIds")
	}

	if p.verifyWebhookMessage == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "verifyWebhookMessage")
	}

	if p.register == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "register")
	}

	if p.deleteRegistration == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "deleteRegistration")
	}

	if p.emptyRegistrationParams == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "emptyRegistrationParams")
	}

	if p.emptyRegistrationResult == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "emptyRegistrationResult")
	}

	if p.subscribe == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "subscribe")
	}

	if p.updateSubscription == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "updateSubscription")
	}

	if p.deleteSubscription == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "deleteSubscription")
	}

	if p.emptySubscriptionParams == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "emptySubscriptionParams")
	}

	if p.emptySubscriptionResult == nil {
		return fmt.Errorf("%w: %s", ErrMissingParam, "emptySubscriptionResult")
	}

	return nil
}
