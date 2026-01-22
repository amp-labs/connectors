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

// RecordCountConnector is an interface that extends the Connector interface with
// the ability to retrieve record counts.
type RecordCountConnector interface {
	Connector

	// GetRecordCount returns the count of records for the given object and time range.
	//
	// Parameters:
	//   - ctx: context for the operation
	//   - params: parameters specifying the object name and optional time range
	//
	// Returns:
	//   - *RecordCountResult: the result containing the count
	//   - error: any error that occurred while fetching the count
	GetRecordCount(ctx context.Context, params *common.RecordCountParams) (*common.RecordCountResult, error)
}

// BatchRecordReaderConnector defines the interface for connectors that can
// fetch full record data from a provider in batch.
type BatchRecordReaderConnector interface {
	Connector

	// GetRecordsByIds fetches full records from the provider for a specific set of IDs.
	//
	// This method is primarily used during webhook processing to enrich events that
	// require fetching the current full state of records from the provider API.
	// In that lifecycle, webhook payloads often contain only partial data, making
	// an explicit read necessary to produce complete, descriptive ReadResultRow values.
	//
	// More generally, this method represents a targeted read operation: it allows
	// callers to retrieve a known set of records by ID without performing a full
	// collection read. This is useful in non-webhook lifecycles as well, such as
	// enriching read results with associated sub-objects, resolving references to
	// related objects, or performing joined reads where supported by the provider.
	//
	// Compared to Read (which lists collections of records), GetRecordsByIds is a
	// singular-by-identity read that operates over multiple explicit IDs.
	//
	// The connector should:
	//   - Fetch only the specified recordIds
	//   - Respect requested fields and associations when supported by the provider
	//   - Return provider responses translated into ReadResultRow
	GetRecordsByIds(
		ctx context.Context,
		objectName string,
		recordIds []string, //nolint:revive
		fields []string,
		associations []string) ([]common.ReadResultRow, error)
}

// WebhookVerifierConnector defines the interface for connectors that can
// authenticate and process incoming webhook requests from a provider.
//
// Implementations are responsible for verifying that a webhook request
// genuinely originated from the provider.
type WebhookVerifierConnector interface {
	Connector
	BatchRecordReaderConnector

	// VerifyWebhookMessage validates the authenticity of an incoming webhook request.
	//
	// The method should verify that the HTTP request was sent by the provider and
	// has not been tampered with. Verification is provider-specific and may rely on
	// request headers, the raw request body, the request URL, or other metadata.
	//
	// Example: verifying a webhook signature using a shared secret.
	//
	// Returning true allows webhook processing to continue.
	// Returning false indicates the request is not trusted and should be rejected.
	// An error should be returned only for unexpected verification failures.
	//
	// Parameters:
	//   - request: the raw webhook HTTP request received from the provider.
	//   - params: provider-specific and user-specific verification parameters, such as
	//     secrets or configuration needed to validate the webhook signature.
	VerifyWebhookMessage(
		ctx context.Context,
		request *common.WebhookRequest,
		params *common.VerificationParams,
	) (bool, error)
}

// SubscribeConnector defines the interface for connectors that manage webhook subscriptions.
//
// Connectors implementing this interface are responsible for creating, updating, and deleting
// webhook subscriptions in a provider system. The interface extends WebhookVerifierConnector,
// so implementing connectors must also be able to verify incoming webhook requests.
type SubscribeConnector interface {
	WebhookVerifierConnector

	// Subscribe creates webhook subscriptions in the provider.
	//
	// SubscribeParams describe the desired subscription state in a normalized,
	// provider-agnostic format, such as subscribing to objects, certain event types,
	// or specific fields. The connector translates this configuration into
	// provider-specific API calls and returns the resulting subscription state.
	Subscribe(
		ctx context.Context,
		params common.SubscribeParams,
	) (*common.SubscriptionResult, error)

	// UpdateSubscription applies detected changes to an existing provider-side subscription.
	//
	// This method is called only after the framework detects changes in the desired
	// subscription configuration (e.g., objects or events added or removed).
	//
	// The params argument represents the new desired subscription state.
	// The previousResult contains the last known actual subscription state stored
	// by the framework.
	//
	// The connector must apply the necessary provider-specific operations to reconcile
	// the existing subscription with the desired state. The reconciliation process
	// is provider-specific.
	//
	// The returned SubscriptionResult must reflect the actual subscription state
	// after the update and will be persisted for future updates or deletion.
	UpdateSubscription(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)

	// DeleteSubscription removes an existing provider-side subscription.
	//
	// This method is called when the framework determines that no subscription
	// lookups remain (i.e., no objects or events are left to subscribe to).
	//
	// The previousResult contains the provider-specific information needed to
	// identify and delete the subscription resources created by Subscribe or
	// UpdateSubscription.
	//
	// After this method succeeds, the provider should no longer send webhook events
	// for this subscription.
	DeleteSubscription(
		ctx context.Context,
		previousResult common.SubscriptionResult,
	) error

	// EmptySubscriptionParams returns an empty, provider-specific common.SubscribeParams instance.
	//
	// The returned common.SubscribeParams has the Request field (of type any)
	// initialized to an empty value appropriate for the provider.
	EmptySubscriptionParams() *common.SubscribeParams

	// EmptySubscriptionResult returns an empty, provider-specific common.SubscriptionResult instance.
	//
	// The returned common.SubscriptionResult has the Result field (of type any)
	// initialized to an empty value appropriate for storing the provider's raw response.
	EmptySubscriptionResult() *common.SubscriptionResult
}

type RegisterSubscribeConnector interface {
	SubscribeConnector

	// Register performs a provider-specific registration required to enable
	// webhook subscriptions.
	//
	// This is typically a one-time operation per installation that
	// may create shared infrastructure used by all subsequent subscriptions.
	Register(
		ctx context.Context,
		params common.SubscriptionRegistrationParams,
	) (*common.RegistrationResult, error)

	// TODO: Uncomment when we implement UpdateRegistration
	// UpdateRegistration(
	// 	ctx context.Context,
	// 	params SubscriptionRegistrationParams,
	// 	previousResult RegistrationResult,
	// ) (*RegistrationResult, error)

	// DeleteRegistration removes a previously created registration from the provider.
	//
	// This method is called when the framework determines that the registration
	// is no longer needed (for example, when all subscriptions have been removed)
	DeleteRegistration(
		ctx context.Context,
		previousResult common.RegistrationResult,
	) error

	// EmptyRegistrationParams returns an empty, provider-specific common.SubscriptionRegistrationParams instance.
	//
	// The returned common.SubscriptionRegistrationParams has the Request field (of type any)
	// initialized to an empty value appropriate for the provider.
	EmptyRegistrationParams() *common.SubscriptionRegistrationParams

	// EmptyRegistrationResult returns an empty, provider-specific common.RegistrationResult instance.
	//
	// The returned common.RegistrationResult has the Result field (of type any)
	// initialized to an empty value appropriate for storing the provider's raw response.
	EmptyRegistrationResult() *common.RegistrationResult
}

// SubscriptionMaintainerConnector defines the interface for connectors that
// require periodic maintenance to keep subscriptions active.
//
// Some providers issue webhook subscriptions that expire after a fixed time
// and must be periodically renewed or refreshed to remain valid. Connectors
// implementing this interface are responsible for performing any scheduled
// maintenance operations required to prevent subscription expiration.
type SubscriptionMaintainerConnector interface {
	SubscribeConnector

	// RunScheduledMaintenance performs provider-specific maintenance for an
	// existing subscription.
	//
	// The params argument represents the desired subscription state and is
	// typically identical to the currently active configuration.
	//
	// The previousResult contains the last known actual subscription state stored
	// by the framework and may include provider-specific identifiers, timestamps,
	// or expiration information required to renew the subscription.
	//
	// The returned SubscriptionResult must reflect the actual subscription state
	// after maintenance and will be persisted for future maintenance, updates,
	// or deletion.
	RunScheduledMaintenance(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)
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
	WriteType                = common.WriteType
	BatchWriteResult         = common.BatchWriteResult
	BatchStatus              = common.BatchStatus
	ListObjectMetadataResult = common.ListObjectMetadataResult
	RecordCountParams        = common.RecordCountParams
	RecordCountResult        = common.RecordCountResult

	ErrorWithStatus = common.HTTPError //nolint:errname
)

const (
	BatchStatusSuccess = common.BatchStatusSuccess
	BatchStatusFailure = common.BatchStatusFailure
	BatchStatusPartial = common.BatchStatusPartial
	WriteTypeCreate    = common.WriteTypeCreate
	WriteTypeUpdate    = common.WriteTypeUpdate
	WriteTypeDelete    = common.WriteTypeDelete
	WriteTypeUpsert    = common.WriteTypeUpsert
)

var Fields = datautils.NewStringSet // nolint:gochecknoglobals
