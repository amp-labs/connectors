package subscribe

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// PostProcessConfig is the subcomponent of ProviderConfig that exposes the post-subscribe
// third-party setup step some providers require after the connector's Subscribe call returns
// (e.g. Salesforce → AWS EventBridge wiring).
//
// Only the derived "should perform?" answer lives here (from ProviderInfo.SubscribeRequirements.
// PostProcess). The setup/teardown functions themselves are Ampersand infrastructure concerns
// (AWS wiring, server-side registration records) and remain declared server-side — they are not
// part of this library.
//
// module / providerInfo are bound at GetProviderConfig time and threaded into ShouldPerform. The
// zero PostProcessConfig is valid; ProviderConfig embeds it by value, and the method takes a
// value receiver, so callers can invoke it directly without nil-checking — the type system
// guarantees a non-nil receiver.
type PostProcessConfig struct {
	// module and providerInfo are bound by GetProviderConfig at call time. ShouldPerform reads
	// them to compute the answer from ProviderInfo.
	module       common.ModuleID
	providerInfo *providers.ProviderInfo
}

// ShouldPerform reports whether the provider requires a post-subscribe setup step. Derived from
// ProviderInfo via ShouldPostProcess.
func (p PostProcessConfig) ShouldPerform() bool {
	return ShouldPostProcess(p.module, p.providerInfo)
}
