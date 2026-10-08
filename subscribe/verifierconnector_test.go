package subscribe

import (
	"context"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

// TestVerifierConnectorsAnswerProviderAndString pins a production panic: the server wraps the
// verifier connector for metrics before running any verification, and the wrapper's first act is
// to call Provider() on it. Several providers register a bare connector as their verifier — one
// with no embedded *components.Connector — and the base's Provider and String have value
// receivers, so the promoted calls dereferenced that nil pointer and panicked. The Slack message
// events that triggered it were nacked and redelivered forever.
//
// Every registered verifier connector must answer both without panicking. This walks the whole
// registry rather than naming providers, so a new provider registering a bare connector fails
// here instead of in production.
func TestVerifierConnectorsAnswerProviderAndString(t *testing.T) {
	t.Parallel()

	for provider, registry := range providerConfigs {
		for moduleID, cfg := range modulesOf(registry) {
			if cfg == nil || cfg.Verification.verifierConnector == nil {
				continue
			}

			name := string(provider)
			if moduleID != "" {
				name += "/" + string(moduleID)
			}

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				conn, err := cfg.Verification.Connector(context.Background())
				if err != nil {
					t.Fatalf("Connector() error = %v, want nil", err)
				}

				// Both calls panicked for connectors whose base is nil.
				if got := conn.Provider(); got == "" {
					t.Error("Provider() = empty, want the provider name")
				}

				// The identifier must name the connector, not invent a type: every
				// provider that declares its own String returns "<provider>.Connector".
				if got := conn.String(); !strings.HasPrefix(got, conn.Provider()) {
					t.Errorf("String() = %q, want a string starting with %q", got, conn.Provider())
				}
			})
		}
	}
}

// modulesOf flattens a registry entry into module ID -> config, using the empty module ID for a
// provider's default config.
func modulesOf(registry ProviderConfigRegistry) map[common.ModuleID]*ProviderConfig {
	out := make(map[common.ModuleID]*ProviderConfig, len(registry.Modules)+1)

	if registry.DefaultModuleConfig != nil {
		out[""] = registry.DefaultModuleConfig
	}

	for moduleID, cfg := range registry.Modules {
		out[moduleID] = cfg
	}

	return out
}

// TestSlackVerifierConnectorProvider pins the specific connector from the production panic: the
// Slack verifier connector is built with a webhook verifier and no base connector.
func TestSlackVerifierConnectorProvider(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Slack, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	conn, err := cfg.Verification.Connector(context.Background())
	if err != nil {
		t.Fatalf("Connector() error = %v, want nil", err)
	}

	if got := conn.Provider(); got != providers.Slack {
		t.Errorf("Provider() = %q, want %q", got, providers.Slack)
	}
}
