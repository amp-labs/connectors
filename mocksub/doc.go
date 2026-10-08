// Package mocksub provides a generic subscribe-testing connector for the mock providers
// (providers.MockSalesloft, ...) used by subscribe regression tests.
//
// Each mock provider mimics a real provider's webhook-receive behavior: its subscribe config
// (subscribe/mock<provider>.go) reuses the real provider's SubscriptionEvent types and casters,
// so event parsing and classification run the real code, while this connector stands in for the
// HTTP-touching connector methods the receive path needs — GetRecordsByIds serves canned records
// from an in-process Store instead of calling a provider API, and webhook signature verification
// is a happy-path no-op (the mock configs declare verification bypassed).
//
// A test seeds records through the provider's process-wide store before running the flow under
// test:
//
//	providers.SetupMockSalesloftProvider()
//	mocksub.StoreFor(providers.MockSalesloft).Seed("people", "42", map[string]any{"id": 42, ...})
//
// The connector constructed by the connector.New factory for a mock provider reads from that
// same store.
package mocksub
