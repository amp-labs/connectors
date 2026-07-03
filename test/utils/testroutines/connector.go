// Package testroutines holds a collection of common test procedures.
// They provide a framework to write mock tests.
package testroutines

import (
	"context"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
)

// ConnectorBuilder is a callback method to construct and configure connector for testing.
// This is a factory method called for every test suite.
type ConnectorBuilder[Conn any] func() (Conn, error)

func (builder ConnectorBuilder[C]) Build(t *testing.T, testCaseName string) C {
	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", testCaseName, err)
	}

	return conn
}

// TestablePostAuthMetadata is the minimal interface for a connector that returns post auth metadata.
type TestablePostAuthMetadata interface {
	GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error)
}

// TestableMetadataReader is the minimal interface for a connector that can read metadata.
type TestableMetadataReader interface {
	ListObjectMetadata(ctx context.Context, objectNames []string) (*common.ListObjectMetadataResult, error)
}

// TestableMetadataDeleter is the minimal interface for a connector that can delete metadata.
type TestableMetadataDeleter interface {
	DeleteMetadata(ctx context.Context, params *common.DeleteMetadataParams) (*common.DeleteMetadataResult, error)
}

// TestableReader is the minimal interface for a connector that can read records.
type TestableReader interface {
	Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error)
}

// TestableDeleter is the minimal interface for a connector that can delete records.
type TestableDeleter interface {
	Delete(ctx context.Context, params common.DeleteParams) (*common.DeleteResult, error)
}

// TestableBatchReader is the minimal interface for a connector that can batch read records.
type TestableBatchReader interface {
	GetRecordsByIds(
		ctx context.Context,
		objectName string,
		recordIds []string,
		fields []string,
		associations []string,
	) ([]common.ReadResultRow, error)
}

// TestableBatchWriter is the minimal interface for a connector that can batch write records.
type TestableBatchWriter interface {
	BatchWrite(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error)
}

// TestableWebhookMessageVerifier is the minimal interface for a connector that can verify webhook messages.
type TestableWebhookMessageVerifier interface {
	VerifyWebhookMessage(
		ctx context.Context,
		request *common.WebhookRequest,
		params *common.VerificationParams,
	) (bool, error)
}

// TestableSubscriptionCreator is the minimal interface for a connector that can create subscriptions.
type TestableSubscriptionCreator interface {
	Subscribe(
		ctx context.Context,
		params common.SubscribeParams,
	) (*common.SubscriptionResult, error)
}

// TestableSubscriptionUpdater is the minimal interface for a connector that can update subscriptions.
type TestableSubscriptionUpdater interface {
	UpdateSubscription(
		ctx context.Context,
		params common.SubscribeParams,
		previousResult *common.SubscriptionResult,
	) (*common.SubscriptionResult, error)
}

// TestableSubscriptionRemover is the minimal interface for a connector that can delete subscriptions.
type TestableSubscriptionRemover interface {
	DeleteSubscription(
		ctx context.Context,
		previousResult common.SubscriptionResult,
	) error
}

// Compile-time assertion that the minimal subscription interfaces
// satisfy connectors.SubscribeConnector.
//
// Each connector asserts only the interfaces it implements.
// If connectors.SubscribeConnector changes, the compiler forces updates,
// causing dependent connectors to fail in tests.
//
// Enables incremental, method-by-method implementation with full compatibility and testing.
var (
	_ connectors.SubscribeConnector = (*dummySubscribeConnector)(nil)
)

// dummySubscribeConnector composes the minimal subscription interfaces and
// required base behavior. It has no implementations and exists purely for
// compile-time interface verification.
type dummySubscribeConnector struct {
	// Base.
	connectors.BatchRecordReaderConnector

	// Decomposed interfaces (primary).
	TestableWebhookMessageVerifier
	TestableSubscriptionCreator
	TestableSubscriptionUpdater
	TestableSubscriptionRemover

	// Supporting helpers (secondary).
	components.SubscriptionInputOutput[any, any]
}
