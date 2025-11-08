package connectors

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
)

// Connector is an interface that can be used to implement a connector with
// basic configuration about the provider.
type Connector interface {
	fmt.Stringer

	// JSONHTTPClient returns the underlying JSON HTTP client. This is useful for
	// testing, or for calling methods that aren't exposed by the Connector
	// interface directly. Authentication and token refreshes will be handled automatically.
	JSONHTTPClient() *common.JSONHTTPClient

	// HTTPClient returns the underlying HTTP client. This is useful for proxy requests.
	HTTPClient() *common.HTTPClient

	// Provider returns the connector provider.
	Provider() providers.Provider
}

// URLConnector is an interface that extends the Connector interface with the ability to
// retrieve URLs for resources.
type URLConnector interface {
	Connector

	// GetURL returns the URL of some resource. The resource is provider-specific.
	// The URL is returned as a string, or an error is returned if the URL cannot be
	// retrieved. The precise meaning of the resource is provider-specific, and the
	// caller should consult the provider's documentation for more information.
	// The args parameter is a map of key-value pairs that can be used to customize
	// the URL. The keys and values are provider-specific, and the caller should
	// consult the provider's documentation for more information. Certain providers
	// may ignore the args parameter entirely if it's unnecessary.
	GetURL(resource string, args map[string]any) (string, error)
}

// ReadConnector is an interface that extends the Connector interface with read capabilities.
type ReadConnector interface {
	Connector

	// Read reads a page of data from the connector. This can be called multiple
	// times to read all the data. The caller is responsible for paging, by
	// passing the NextPage value correctly, and by terminating the loop when
	// Done is true. The caller is also responsible for handling errors.
	// Authentication corner cases are handled internally, but all other errors
	// are returned to the caller.
	Read(ctx context.Context, params ReadParams) (*ReadResult, error)
}

// WriteConnector is an interface that extends the Connector interface with write capabilities.
type WriteConnector interface {
	Connector

	Write(ctx context.Context, params WriteParams) (*WriteResult, error)
}

// DeleteConnector is an interface that extends the Connector interface with delete capabilities.
type DeleteConnector interface {
	Connector

	Delete(ctx context.Context, params DeleteParams) (*DeleteResult, error)
}

// BatchWriteConnector provides synchronous operations for writing multiple records in a single request.
// It serves the same purpose as WriteConnector but operates
// on collections of records instead of individual ones.
//
// Implementations should handle each record independently and report both
// overall and per-record outcomes through the returned result types.
// Errors returned from the methods represent connector-level issues such as
// network failures or invalid authentication, not individual record failures.
type BatchWriteConnector interface {
	Connector

	// BatchWrite performs a batch create, update, or upsert operation.
	// Each record in params.Records is processed according to params.Type.
	// The returned BatchWriteResult includes both per-record outcomes and the aggregate batch status.
	BatchWrite(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error)
}

// ObjectMetadataConnector is an interface that extends the Connector interface with
// the ability to list object metadata.
type ObjectMetadataConnector interface {
	Connector

	ListObjectMetadata(ctx context.Context, objectNames []string) (*ListObjectMetadataResult, error)
}

// UpsertMetadataConnector is an interface that extends the Connector interface with
// the ability to create/update custom objects and fields in the SaaS instance.
type UpsertMetadataConnector interface {
	Connector

	UpsertMetadata(ctx context.Context, params *common.UpsertMetadataParams) (*common.UpsertMetadataResult, error)
}

// AuthMetadataConnector is an interface that extends the Connector interface with
// the ability to retrieve metadata information about authentication.
type AuthMetadataConnector interface {
	Connector

	// GetPostAuthInfo returns authentication metadata.
	GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error)
}

type BatchRecordReaderConnector interface {
	Connector
	GetRecordsByIds(
		ctx context.Context,
		objectName string,
		//nolint:revive
		recordIds []string,
		fields []string,
		associations []string) ([]common.ReadResultRow, error)
}
type WebhookVerifierConnector interface {
	Connector
	BatchRecordReaderConnector

	// VerifyWebhookMessage verifies the signature of a webhook message.
	VerifyWebhookMessage(
		ctx context.Context,
		// request is the raw webhook request from the provider.
		request *common.WebhookRequest,
		// params is the verification parameters unique to the user.
		// It is used to verify the signature of the webhook message.
		params *common.VerificationParams,
	) (bool, error)
}

type RegisterSubscribeConnector interface {
	// RegisterSubscribeConnector has below responsibilities:
	// 1. Register a subscription with the provider.
	// Registering a subscription is a one-time operation that is required
	// by providers that hold some master registration of all subscriptions.
	// Not all providers require this, but some do.
	SubscribeConnector
	Register(
		ctx context.Context,
		params common.SubscriptionRegistrationParams,
	) (*common.RegistrationResult, error)
	// TODO: Uncomment when we implement UpdateRegistration in Salesforce
	// UpdateRegistration(
	// 	ctx context.Context,
	// 	params SubscriptionRegistrationParams,
	// 	previousResult RegistrationResult,
	// ) (*RegistrationResult, error)
	DeleteRegistration(
		ctx context.Context,
		previousResult common.RegistrationResult,
	) error
	// EmptyRegistrationParams returns a empty instance of SubscriptionRegistrationParams.
	// if there is any provider specific initialization required, it should be done here.
	EmptyRegistrationParams() *common.SubscriptionRegistrationParams
	// EmptyRegistrationResult returns a empty instance of RegistrationResult.
	// if there is any provider specific initialization required, it should be done here.
	EmptyRegistrationResult() *common.RegistrationResult
}

//nolint:interfacebloat
type SubscribeConnector interface {
	WebhookVerifierConnector
	// SubscribeConnector has below responsibilities:
	// Subscribe to events from the provider.
	// This is the actual subscription to events from the provider.
	// It will subscribe for events and objects as specified in SubscribeParams.
	Subscribe(
		ctx context.Context,
		params common.SubscribeParams,
	) (*common.SubscriptionResult, error)
	UpdateSubscription(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)
	DeleteSubscription(
		ctx context.Context,
		previousResult common.SubscriptionResult,
	) error
	// EmptySubscriptionParams returns a empty instance of SubscribeParams.
	// if there is any provider specific initialization required, it should be done here.
	EmptySubscriptionParams() *common.SubscribeParams
	// EmptySubscriptionResult returns a empty instance of SubscriptionResult.
	// if there is any provider specific initialization required, it should be done here.
	EmptySubscriptionResult() *common.SubscriptionResult
	// GetRecordsWithId is a helper function to get records by their IDs.
	//nolint:revive
}

// ConfigurationConnector is a connector that has methods to expose connector
// configuration values to a caller. This is an interface as opposed to a
// ProviderInfo value because PageSize might change based on the provider license
// or based on the endpoint, in which case we can modify DefaultPageSize() to accept
// ReadParams as well.
type ConfigurationConnector interface {
	Connector

	DefaultPageSize() int
}

// We re-export the following types so that they can be used by consumers of this library.
type (
	ReadParams               = common.ReadParams
	WriteParams              = common.WriteParams
	DeleteParams             = common.DeleteParams
	ReadResult               = common.ReadResult
	WriteResult              = common.WriteResult
	DeleteResult             = common.DeleteResult
	BatchWriteParam          = common.BatchWriteParam
	BatchWriteType           = common.BatchWriteType
	BatchWriteResult         = common.BatchWriteResult
	BatchStatus              = common.BatchStatus
	ListObjectMetadataResult = common.ListObjectMetadataResult

	ErrorWithStatus = common.HTTPError //nolint:errname
)

const (
	BatchStatusSuccess   = common.BatchStatusSuccess
	BatchStatusFailure   = common.BatchStatusFailure
	BatchStatusPartial   = common.BatchStatusPartial
	BatchWriteTypeCreate = common.BatchWriteTypeCreate
	BatchWriteTypeUpdate = common.BatchWriteTypeUpdate
)

var Fields = datautils.NewStringSet // nolint:gochecknoglobals

type SubscriptionMaintainerConnector interface {
	Connector
	RunScheduledMaintenance(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)
}
