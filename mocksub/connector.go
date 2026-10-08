package mocksub

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// ObjectNameFromEventFunc resolves the object name for a subscription event. Declared for mock
// providers mimicking a provider whose events identify the object indirectly (e.g. Attio's
// record.* events carry a provider-side object id), matching
// connectors.SubscriptionEventObjectNameConnector.
type ObjectNameFromEventFunc func(ctx context.Context, event common.SubscriptionEvent) (string, error)

// Connector is the generic connector behind the mock subscribe providers. The webhook-receive
// path needs only the WebhookVerifierConnector surface; every method answers from in-process
// state, so no HTTP call ever leaves the test.
type Connector struct {
	provider            providers.Provider
	client              *common.JSONHTTPClient
	store               *Store
	objectNameFromEvent ObjectNameFromEventFunc
}

var (
	_ connectors.WebhookVerifierConnector             = (*Connector)(nil)
	_ connectors.SubscriptionEventObjectNameConnector = (*Connector)(nil)
)

// Option configures a Connector.
type Option func(*Connector)

// WithStore overrides the canned-record store. Defaults to the provider's StoreFor singleton,
// which is what the connector.New factory path uses.
func WithStore(store *Store) Option {
	return func(c *Connector) {
		c.store = store
	}
}

// WithObjectNameFromEvent declares the event-to-object-name resolver for mock providers whose
// events identify the object indirectly. When absent, GetObjectNameFromEvent falls back to the
// event's own ObjectName.
func WithObjectNameFromEvent(fn ObjectNameFromEventFunc) Option {
	return func(c *Connector) {
		c.objectNameFromEvent = fn
	}
}

// NewConnector returns a mock connector answering for the given provider name.
func NewConnector(provider providers.Provider, opts ...Option) *Connector {
	conn := &Connector{
		provider: provider,
		client: &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client: http.DefaultClient,
			},
		},
		store: StoreFor(provider),
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn
}

func (c *Connector) String() string {
	return c.provider
}

func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.client.HTTPClient
}

func (c *Connector) Provider() providers.Provider {
	return c.provider
}

// GetRecordsByIds serves records from the canned store with real-connector semantics: ids with
// no stored record are silently omitted (providers omit unknown ids from batch reads), and rows
// are built by common.GetMarshaledData — the same helper real connectors use — so Raw carries
// the full record, Fields the lowercased requested subset, and Id the record's "id" value.
//
//nolint:revive // recordIds parameter name matches the connectors interface.
func (c *Connector) GetRecordsByIds(
	_ context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	records := make([]map[string]any, 0, len(recordIds))

	for _, id := range recordIds {
		if record, ok := c.store.Get(objectName, id); ok {
			records = append(records, record)
		}
	}

	return common.GetMarshaledData(records, fields)
}

// VerifyWebhookMessage reports every message as valid. Mock providers declare verification
// bypassed in their subscribe configs, so this exists to satisfy the WebhookVerifierConnector
// interface rather than to gate anything; happy-path regression tests never exercise rejection.
func (c *Connector) VerifyWebhookMessage(
	_ context.Context,
	_ *common.WebhookRequest,
	_ *common.VerificationParams,
) (bool, error) {
	return true, nil
}

// GetObjectNameFromEvent resolves the event's object name via the declared resolver, falling
// back to the event's own ObjectName when the mock provider declares none — the same answer
// the receive path computes for connectors that do not implement
// SubscriptionEventObjectNameConnector.
func (c *Connector) GetObjectNameFromEvent(
	ctx context.Context,
	event common.SubscriptionEvent,
) (string, error) {
	if c.objectNameFromEvent != nil {
		return c.objectNameFromEvent(ctx, event)
	}

	return event.ObjectName()
}
