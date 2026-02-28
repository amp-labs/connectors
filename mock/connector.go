package mock

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	client *common.JSONHTTPClient
	params *parameters
}

// We want the mock connector to implement all connector interfaces.
type implementsAllConnector interface {
	connectors.Connector
	connectors.URLConnector
	connectors.ReadConnector
	connectors.WriteConnector
	connectors.DeleteConnector
	connectors.ObjectMetadataConnector
	connectors.AuthMetadataConnector
	connectors.BatchRecordReaderConnector
	connectors.WebhookVerifierConnector
	connectors.SubscribeConnector
}

var _ implementsAllConnector = (*Connector)(nil)

func NewConnector(opts ...Option) (conn *Connector, outErr error) { //nolint:funlen
	params, err := paramsbuilder.Apply(parameters{}, opts,
		WithClient(http.DefaultClient),
		WithRead(func(context.Context, common.ReadParams) (*common.ReadResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "read")
		}),
		WithWrite(func(context.Context, common.WriteParams) (*common.WriteResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "write")
		}),
		WithListObjectMetadata(func(context.Context, []string) (*common.ListObjectMetadataResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "listObjectMetadata")
		}),
		WithGetURL(func(resource string, args map[string]any) (string, error) {
			return "", fmt.Errorf("%w: %s", ErrNotImplemented, "getURL")
		}),
		WithDelete(func(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "delete")
		}),
		WithGetPostAuthInfo(func(ctx context.Context) (*common.PostAuthInfo, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "getPostAuthInfo")
		}),
		WithGetRecordsByIds(func(ctx context.Context, params common.ReadByIdsParams) ([]common.ReadResultRow, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "getRecordsByIds")
		}),
		WithVerifyWebhookMessage(
			func(
				ctx context.Context,
				request *common.WebhookRequest,
				params *common.VerificationParams,
			) (bool, error) {
				return false, fmt.Errorf("%w: %s", ErrNotImplemented, "verifyWebhookMessage")
			}),
		WithRegister(func(
			ctx context.Context,
			params common.SubscriptionRegistrationParams,
		) (*common.RegistrationResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "register")
		}),
		WithDeleteRegistration(func(ctx context.Context, previousResult common.RegistrationResult) error {
			return fmt.Errorf("%w: %s", ErrNotImplemented, "deleteRegistration")
		}),
		WithEmptyRegistrationParams(func() *common.SubscriptionRegistrationParams {
			return &common.SubscriptionRegistrationParams{
				Request: make(RegistrationRequest),
			}
		}),
		WithEmptyRegistrationResult(func() *common.RegistrationResult {
			return &common.RegistrationResult{
				Result: make(RegistrationResult),
			}
		}),
		WithSubscribe(func(ctx context.Context, params common.SubscribeParams) (*common.SubscriptionResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "subscribe")
		}),
		WithUpdateSubscription(func(
			ctx context.Context,
			params common.SubscribeParams,
			previousResult *common.SubscriptionResult,
		) (*common.SubscriptionResult, error) {
			return nil, fmt.Errorf("%w: %s", ErrNotImplemented, "updateSubscription")
		}),
		WithDeleteSubscription(func(ctx context.Context, previousResult common.SubscriptionResult) error {
			return fmt.Errorf("%w: %s", ErrNotImplemented, "deleteSubscription")
		}),
		WithEmptySubscriptionParams(func() *common.SubscribeParams {
			return &common.SubscribeParams{
				Request: make(SubscriptionRequest),
			}
		}),
		WithEmptySubscriptionResult(func() *common.SubscriptionResult {
			return &common.SubscriptionResult{
				Result: make(SubscriptionResult),
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Connector{
		client: params.client,
		params: params,
	}, nil
}

func (c *Connector) String() string {
	return "mock"
}

func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.client.HTTPClient
}

func (c *Connector) Provider() providers.Provider {
	return providers.Mock
}

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	return c.params.read(ctx, params)
}

func (c *Connector) Write(ctx context.Context, params common.WriteParams) (*common.WriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return c.params.write(ctx, params)
}

func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	return c.params.listObjectMetadata(ctx, objectNames)
}

func (c *Connector) GetURL(resource string, args map[string]any) (string, error) {
	return c.params.getURL(resource, args)
}

func (c *Connector) Delete(ctx context.Context, params connectors.DeleteParams) (*connectors.DeleteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	return c.params.delete(ctx, params)
}

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	return c.params.getPostAuthInfo(ctx)
}

//nolint:revive
func (c *Connector) GetRecordsByIds(ctx context.Context,
	params common.ReadByIdsParams,
) ([]common.ReadResultRow, error) {
	return c.params.getRecordsByIds(ctx, params)
}

func (c *Connector) VerifyWebhookMessage(
	ctx context.Context,
	request *common.WebhookRequest,
	params *common.VerificationParams,
) (bool, error) {
	return c.params.verifyWebhookMessage(ctx, request, params)
}

func (c *Connector) Register(
	ctx context.Context,
	params common.SubscriptionRegistrationParams,
) (*common.RegistrationResult, error) {
	return c.params.register(ctx, params)
}

func (c *Connector) DeleteRegistration(ctx context.Context, previousResult common.RegistrationResult) error {
	return c.params.deleteRegistration(ctx, previousResult)
}

func (c *Connector) EmptyRegistrationParams() *common.SubscriptionRegistrationParams {
	return c.params.emptyRegistrationParams()
}

func (c *Connector) EmptyRegistrationResult() *common.RegistrationResult {
	return c.params.emptyRegistrationResult()
}

func (c *Connector) Subscribe(ctx context.Context, params common.SubscribeParams) (*common.SubscriptionResult, error) {
	return c.params.subscribe(ctx, params)
}

func (c *Connector) UpdateSubscription(
	ctx context.Context,
	params common.SubscribeParams,
	previousResult *common.SubscriptionResult,
) (*common.SubscriptionResult, error) {
	return c.params.updateSubscription(ctx, params, previousResult)
}

func (c *Connector) DeleteSubscription(ctx context.Context, previousResult common.SubscriptionResult) error {
	return c.params.deleteSubscription(ctx, previousResult)
}

func (c *Connector) EmptySubscriptionParams() *common.SubscribeParams {
	return c.params.emptySubscriptionParams()
}

func (c *Connector) EmptySubscriptionResult() *common.SubscriptionResult {
	return c.params.emptySubscriptionResult()
}
