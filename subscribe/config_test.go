package subscribe

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// makeProviderInfo loads the provider's real ProviderInfo from the providers catalog so the
// ProviderInfo-derived subscribe methods (IsSupportedViaAPI, SubscribeManually, ShouldPerform, …)
// resolve against the actual Support / SubscribeRequirements data. When module is non-empty it
// overrides DefaultModule so an empty-module lookup resolves to that module. Unknown providers
// (the not-found test case) fall back to a minimal stub. ModuleID is a string alias, so the
// module constants pass through as plain strings.
func makeProviderInfo(name providers.Provider, module string) *providers.ProviderInfo {
	info, err := providers.ReadInfo(name)
	if err != nil {
		return &providers.ProviderInfo{Name: name, DefaultModule: module}
	}

	if module != "" {
		infoCopy := *info
		infoCopy.DefaultModule = module
		info = &infoCopy
	}

	return ResolveProviderInfoAlias(info)
}

// TestResolveModuleSubscribeDataModuleSelection pins the module-selection rule: a non-empty
// module is the caller's explicit choice and is the only module consulted; only when the module
// is empty does providerInfo.DefaultModule come into play. In particular, an unknown module must
// NOT silently fall back to DefaultModule — that would change the requested semantics.
func TestResolveModuleSubscribeDataModuleSelection(t *testing.T) {
	t.Parallel()

	modules := providers.Modules{
		providers.ModuleSalesforceCRM: providers.ModuleInfo{
			Support: providers.Support{Subscribe: true},
			SubscribeRequirements: &providers.SubscribeRequirements{
				Registration:   new(true),
				SubscribeByAPI: new(true),
			},
		},
	}

	topLevelReqs := &providers.SubscribeRequirements{
		Registration:   new(false),
		SubscribeByAPI: new(false),
	}

	provInfo := &providers.ProviderInfo{
		Name:                  providers.Salesforce,
		Support:               providers.Support{Subscribe: false},
		Modules:               &modules,
		DefaultModule:         providers.ModuleSalesforceCRM,
		SubscribeRequirements: topLevelReqs,
	}

	tests := []struct {
		name             string
		module           common.ModuleID
		wantSubscribe    bool
		wantRegistration bool
	}{
		{
			name:             "module matches registered module → use module values",
			module:           providers.ModuleSalesforceCRM,
			wantSubscribe:    true,
			wantRegistration: true,
		},
		{
			name:             "empty module → use DefaultModule → module values",
			module:           "",
			wantSubscribe:    true,
			wantRegistration: true,
		},
		{
			// Regression guard: a non-empty module that isn't registered must NOT silently
			// fall back to DefaultModule. The lookup misses and the top-level Support /
			// SubscribeRequirements are used instead.
			name:             "unknown module → top-level (DefaultModule NOT consulted)",
			module:           "unknown-module",
			wantSubscribe:    false,
			wantRegistration: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			support, reqs := resolveModuleSubscribeData(tc.module, provInfo)
			if support.Subscribe != tc.wantSubscribe {
				t.Errorf("Support.Subscribe: got %v, want %v", support.Subscribe, tc.wantSubscribe)
			}

			if reqs == nil || reqs.Registration == nil {
				t.Fatalf("expected non-nil Registration, got %+v", reqs)
			}

			if *reqs.Registration != tc.wantRegistration {
				t.Errorf("Registration: got %v, want %v", *reqs.Registration, tc.wantRegistration)
			}
		})
	}
}

// TestResolveModuleSubscribeDataNoMatch covers the case where neither the supplied module nor
// DefaultModule is registered in providerInfo.Modules — the helper must fall back to top-level
// Support / SubscribeRequirements.
func TestResolveModuleSubscribeDataNoMatch(t *testing.T) {
	t.Parallel()

	modules := providers.Modules{
		"only-this-one": providers.ModuleInfo{
			Support:               providers.Support{Subscribe: true},
			SubscribeRequirements: &providers.SubscribeRequirements{SubscribeByAPI: new(true)},
		},
	}

	topLevelReqs := &providers.SubscribeRequirements{SubscribeByAPI: new(false)}
	provInfo := &providers.ProviderInfo{
		Name:                  providers.Hubspot,
		Support:               providers.Support{Subscribe: false},
		Modules:               &modules,
		DefaultModule:         "also-unknown",
		SubscribeRequirements: topLevelReqs,
	}

	support, reqs := resolveModuleSubscribeData("unrelated", provInfo)
	if support.Subscribe {
		t.Errorf("Support.Subscribe: got true, want false (top-level)")
	}

	if reqs != topLevelReqs {
		t.Errorf("expected top-level SubscribeRequirements, got %+v", reqs)
	}
}

// TestResolveModuleSubscribeDataNilModules covers providers that declare no Modules at all —
// must use top-level Support / SubscribeRequirements directly.
func TestResolveModuleSubscribeDataNilModules(t *testing.T) {
	t.Parallel()

	topLevelReqs := &providers.SubscribeRequirements{Registration: new(true)}
	provInfo := &providers.ProviderInfo{
		Name:                  providers.Outreach,
		Support:               providers.Support{Subscribe: true},
		Modules:               nil,
		DefaultModule:         "",
		SubscribeRequirements: topLevelReqs,
	}

	support, reqs := resolveModuleSubscribeData("anything", provInfo)
	if !support.Subscribe {
		t.Errorf("Support.Subscribe: got false, want true (top-level)")
	}

	if reqs != topLevelReqs {
		t.Errorf("expected top-level SubscribeRequirements, got %+v", reqs)
	}
}

// TestResolveModuleSubscribeDataModuleNilRequirements covers modules whose ModuleInfo declares no
// SubscribeRequirements — the helper falls back to top-level SubscribeRequirements (the module's
// Support still applies).
func TestResolveModuleSubscribeDataModuleNilRequirements(t *testing.T) {
	t.Parallel()

	modules := providers.Modules{
		"crm": providers.ModuleInfo{
			Support:               providers.Support{Subscribe: true},
			SubscribeRequirements: nil,
		},
	}

	topLevelReqs := &providers.SubscribeRequirements{Registration: new(true)}
	provInfo := &providers.ProviderInfo{
		Support:               providers.Support{Subscribe: false},
		Modules:               &modules,
		DefaultModule:         "crm",
		SubscribeRequirements: topLevelReqs,
	}

	support, reqs := resolveModuleSubscribeData("", provInfo)
	if !support.Subscribe {
		t.Errorf("Support.Subscribe: got false, want true (from module)")
	}

	if reqs != topLevelReqs {
		t.Errorf("expected top-level SubscribeRequirements (module declared nil), got %+v", reqs)
	}
}

func TestGetProviderConfigResolution(t *testing.T) {
	t.Parallel()

	t.Run("module-specific entry resolves", func(t *testing.T) {
		t.Parallel()

		cfg, err := GetProviderConfig("", makeProviderInfo(providers.Salesforce, providers.ModuleSalesforceCRM), deps.Dependencies{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg == nil {
			t.Fatal("expected a config, got nil")
		}
	})

	t.Run("default-module entry resolves", func(t *testing.T) {
		t.Parallel()

		cfg, err := GetProviderConfig("", makeProviderInfo(providers.Outreach, ""), deps.Dependencies{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg == nil {
			t.Fatal("expected a config, got nil")
		}
	})

	t.Run("unregistered provider returns ErrProviderConfigNotFound", func(t *testing.T) {
		t.Parallel()

		_, err := GetProviderConfig("", makeProviderInfo(providers.Provider("nonexistent-provider"), ""), deps.Dependencies{})
		if !errors.Is(err, ErrProviderConfigNotFound) {
			t.Errorf("got %v, want ErrProviderConfigNotFound", err)
		}
	})

	t.Run("nil providerInfo returns ErrProviderConfigNotFound", func(t *testing.T) {
		t.Parallel()

		_, err := GetProviderConfig("", nil, deps.Dependencies{})
		if !errors.Is(err, ErrProviderConfigNotFound) {
			t.Errorf("got %v, want ErrProviderConfigNotFound", err)
		}
	})
}

// TestProviderConfigSubscribeMethods covers the subcomponent methods the subscribe-install call
// sites rely on: Subscription.IsSupportedViaAPI, Subscription.SubscribeManually, and
// Maintenance.Interval — resolved against the real providers catalog.
func TestProviderConfigSubscribeMethods(t *testing.T) {
	t.Parallel()

	//nolint:mnd,lll
	tests := []struct {
		name                  string
		provider              providers.Provider
		module                string
		wantViaAPI            bool
		wantSubscribeManually bool
		wantInterval          time.Duration
		wantHasInterval       bool
		wantPostProcess       bool
		wantWatchFieldsAll    bool
		wantBypass            bool
	}{
		{"salesforce", providers.Salesforce, providers.ModuleSalesforceCRM, true, false, 0, false, true, false, false},
		{"zoho", providers.Zoho, providers.ModuleZohoCRM, true, false, 7 * 24 * time.Hour, true, false, false, false},
		{"google", providers.Google, providers.ModuleGoogleGmail, true, false, 24 * time.Hour, true, false, false, true},
		{"salesloft", providers.Salesloft, "", true, false, 0, false, false, true, false},
		{"hubspot (look-up-only)", providers.Hubspot, "", false, true, 0, false, false, false, true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := GetProviderConfig(common.ModuleID(testCase.module), makeProviderInfo(testCase.provider, testCase.module), deps.Dependencies{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := cfg.Subscription.IsSupportedViaAPI(); got != testCase.wantViaAPI {
				t.Errorf("Subscription.IsSupportedViaAPI() = %v, want %v", got, testCase.wantViaAPI)
			}

			if got := cfg.Subscription.SubscribeManually(); got != testCase.wantSubscribeManually {
				t.Errorf("Subscription.SubscribeManually() = %v, want %v", got, testCase.wantSubscribeManually)
			}

			interval, ok := cfg.Maintenance.Interval()
			if ok != testCase.wantHasInterval {
				t.Errorf("Maintenance.Interval() ok = %v, want %v", ok, testCase.wantHasInterval)
			}

			if interval != testCase.wantInterval {
				t.Errorf("Maintenance.Interval() = %v, want %v", interval, testCase.wantInterval)
			}

			if got := cfg.PostProcess.ShouldPerform(); got != testCase.wantPostProcess {
				t.Errorf("PostProcess.ShouldPerform() = %v, want %v", got, testCase.wantPostProcess)
			}

			if got := cfg.Subscription.RequiresWatchFieldsAutoAll(); got != testCase.wantWatchFieldsAll {
				t.Errorf("Subscription.RequiresWatchFieldsAutoAll() = %v, want %v", got, testCase.wantWatchFieldsAll)
			}

			if got := cfg.Verification.ShouldBypass(); got != testCase.wantBypass {
				t.Errorf("Verification.ShouldBypass() = %v, want %v", got, testCase.wantBypass)
			}
		})
	}
}

// TestSubscriptionBuildRequestNoBuilder verifies the no-op contract: a provider that declares no
// subscribe-time request builder returns (nil, nil).
func TestSubscriptionBuildRequestNoBuilder(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Hubspot, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, err := cfg.Subscription.BuildRequest(context.Background(), nil, nil, nil, "")
	if err != nil {
		t.Errorf("BuildRequest() error = %v, want nil", err)
	}

	if req != nil {
		t.Errorf("BuildRequest() = %v, want nil", req)
	}
}

// TestSubscriptionBuildRequestNilInstallation verifies that a provider which declares a builder
// returns a clear error (not a panic) when BuildRequest is invoked with a nil installation.
func TestSubscriptionBuildRequestNilInstallation(t *testing.T) {
	t.Parallel()

	cfg, err := GetProviderConfig("", makeProviderInfo(providers.Outreach, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, err := cfg.Subscription.BuildRequest(context.Background(), nil, nil, nil, "")
	if !errors.Is(err, errInstallationNotFound) {
		t.Errorf("BuildRequest() error = %v, want errInstallationNotFound", err)
	}

	if req != nil {
		t.Errorf("BuildRequest() = %v, want nil", req)
	}
}

// TestGetGoogleRequestNilRevision verifies the Gmail request builder returns an error rather than
// panicking when called without a revision (it needs the revision's module).
func TestGetGoogleRequestNilRevision(t *testing.T) {
	t.Parallel()

	req, err := getGoogleRequest(context.Background(), deps.Dependencies{}, nil, nil, nil, nil, "")
	if !errors.Is(err, errNilRevision) {
		t.Errorf("getGoogleRequest() error = %v, want errNilRevision", err)
	}

	if req != nil {
		t.Errorf("getGoogleRequest() = %v, want nil", req)
	}
}

// TestVerificationCastEvents verifies the event-caster contract: a provider that declares a
// caster (Hubspot) casts without error; one that declares none (Salesforce) returns
// errSubscriptionEventCasterNotDeclared.
func TestVerificationCastEvents(t *testing.T) {
	t.Parallel()

	hubspotCfg, err := GetProviderConfig("", makeProviderInfo(providers.Hubspot, ""), deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := hubspotCfg.Verification.CastEvents([]map[string]any{})
	if err != nil {
		t.Errorf("CastEvents() error = %v, want nil", err)
	}

	if len(events) != 0 {
		t.Errorf("CastEvents() len = %d, want 0", len(events))
	}

	sfInfo := makeProviderInfo(providers.Salesforce, providers.ModuleSalesforceCRM)

	sfCfg, err := GetProviderConfig("", sfInfo, deps.Dependencies{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = sfCfg.Verification.CastEvents([]map[string]any{})
	if !errors.Is(err, errSubscriptionEventCasterNotDeclared) {
		t.Errorf("CastEvents() error = %v, want errSubscriptionEventCasterNotDeclared", err)
	}
}

// TestProviderConfigRegistryCompleteness is a guardrail test that prevents
// enabling Subscribe support in the provider catalog without providing the
// corresponding configuration in this package. It ensures that every provider
// or module with Subscribe enabled has either:
//   - A module-specific config in Modules[moduleID], or
//   - A fallback DefaultModuleConfig for module-agnostic providers (e.g., HubSpot).
//
// The test dynamically builds its assertions from the catalog, so adding a new
// provider or enabling Subscribe automatically adds a corresponding check here.
func TestProviderConfigRegistryCompleteness(t *testing.T) {
	providerCatalog, err := providers.ReadCatalog()
	if err != nil {
		t.Fatalf("cannot read provider catalog: %v", err)
	}

	type testCase struct {
		name      string
		assertion *testutils.CompareResult
	}

	// Pre-allocate the maximum possible number of test cases (one per provider).
	// The actual number may be lower if some providers have no Subscribe-enabled modules.
	tests := make([]testCase, len(providerCatalog.Catalog))
	idx := -1

	for providerName, providerInfo := range providerCatalog.Catalog {
		idx++

		// Build assertions for provider-level Subscribe (module-agnostic case).
		providerAssertion := testutils.NewCompareResult()
		if providerInfo.Support.Subscribe {
			providerAssertion.Merge(assertProviderSubscribeEnabled(providerName))
		}

		// Build assertions for each module with Subscribe enabled.
		moduleAssertion := testutils.NewCompareResult()
		hasSubscribeModules := false
		if providerInfo.Modules != nil {
			for moduleName, moduleInfo := range *providerInfo.Modules {
				if moduleInfo.Support.Subscribe {
					moduleAssertion.Merge(assertModuleSubscribeEnabled(providerName, moduleName))
					hasSubscribeModules = true
				}
			}
		}

		// Both assertions failed, it is worth reporting everything.
		if !providerAssertion.OK && !moduleAssertion.OK {
			assertion := testutils.NewCompareResult()
			assertion.Merge(providerAssertion)
			assertion.Merge(moduleAssertion)
			tests[idx] = testCase{
				name:      "ProviderConfig should match against Catalog",
				assertion: assertion,
			}
			continue
		}

		if hasSubscribeModules {
			// Module-specific configs take precedence.
			// DefaultModuleConfig is optional and can act as fallback.
			tests[idx] = testCase{
				name:      "ProviderConfigRegistry.Modules should match against Catalog",
				assertion: moduleAssertion,
			}
			continue
		}

		// Provider has no modules. Subscribe support requires at least DefaultModuleConfig.
		tests[idx] = testCase{
			name:      "ProviderConfigRegistry.DefaultModuleConfig should match against Catalog",
			assertion: providerAssertion,
		}
	}

	// Execute all assertions as parallel subtests to surface all failures at once.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.assertion.Validate(t, tt.name)
		})
	}
}

// assertProviderSubscribeEnabled validates that a provider with module-agnostic
// Subscribe support has a non-nil DefaultModuleConfig. This applies to providers
// that do not declare module-specific configs.
func assertProviderSubscribeEnabled(providerName string) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	configRegistry, ok := providerConfigs[providerName]
	if !ok {
		return result.AddDiff("providerConfigs[%v] is missing, "+
			"but subscription to provider is enabled", providerName)
	}
	if configRegistry.DefaultModuleConfig == nil {
		return result.AddDiff("providerConfigs[%v].DefaultModuleConfig is nil", providerName)
	}

	return result
}

// assertModuleSubscribeEnabled validates that a module with Subscribe enabled
// has a corresponding config entry. The config may be:
//   - A module-specific entry in Modules[moduleName], or
//   - A fallback DefaultModuleConfig if the provider uses it as a fallback.
//
// This allows providers like HubSpot to enable Subscribe at the module level
// in the catalog while using a shared DefaultModuleConfig.
func assertModuleSubscribeEnabled(providerName string, moduleName string) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	configRegistry, ok := providerConfigs[providerName]
	if !ok {
		return result.AddDiff("providerConfigs[%v] is missing", providerName)
	}

	if configRegistry.Modules == nil {
		// No module-specific configs declared. Fall back to DefaultModuleConfig.
		if configRegistry.DefaultModuleConfig != nil {
			return result
		}

		return result.AddDiff("providerConfigs[%v].Modules is nil", providerName)
	}

	moduleConfig, ok := configRegistry.Modules[moduleName]
	if !ok {
		// Module-specific entry missing. Fall back to DefaultModuleConfig.
		if configRegistry.DefaultModuleConfig != nil {
			return result
		}

		return result.AddDiff("providerConfigs[%v].Modules[%v] is missing, but subscription to module is enabled", providerName, moduleName)
	}
	if moduleConfig == nil {
		// Module-specific entry is nil. Fall back to DefaultModuleConfig.
		if configRegistry.DefaultModuleConfig != nil {
			return result
		}

		return result.AddDiff("providerConfigs[%v].Modules[%v] is nil", providerName, moduleName)
	}

	return result
}

// TestGetObjectTypeSubscribeEventsListCompleteness is a guardrail test that
// ensures every provider with Subscribe support has an event-format definition
// in GetObjectTypeSubscribeEventsList. It prevents enabling Subscribe in the
// catalog without implementing the corresponding switch case.
//
// Providers that intentionally don't define event formats are added to skipProviders.
func TestGetObjectTypeSubscribeEventsListCompleteness(t *testing.T) {
	// Providers excluded from this check. Add a provider here only if there's
	// a documented reason why Subscribe is enabled but event formats are undefined.
	skipProviders := datautils.NewSet(
		// TODO: Google event formats are not yet defined (why?)
		providers.Google,
		// TODO: HubSpot event formats are not yet defined (why?)
		providers.Hubspot,
	)

	providerCatalog, err := providers.ReadCatalog()
	if err != nil {
		t.Fatalf("cannot read provider catalog: %v", err)
	}

	type testCase struct {
		name     string
		provider providers.Provider
	}

	// Collect all providers with Subscribe enabled (at provider or module level).
	var tests []testCase
	for providerName, providerInfo := range providerCatalog.Catalog {
		if skipProviders.Has(providerName) {
			continue
		}

		hasSubscribe := false
		if providerInfo.Support.Subscribe {
			hasSubscribe = true
		}

		if providerInfo.Modules != nil {
			for _, moduleInfo := range *providerInfo.Modules {
				if moduleInfo.Support.Subscribe {
					hasSubscribe = true
					break
				}
			}
		}

		if hasSubscribe {
			tests = append(tests, testCase{
				name:     fmt.Sprintf("GetObjectTypeSubscribeEventsList must define event for %v", providerName),
				provider: providerName,
			})
		}
	}

	// Execute all assertions as parallel subtests to surface all failures at once.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err = GetObjectTypeSubscribeEventsList(tt.provider, nil)
			if errors.Is(err, errUnsupportedProvider) {
				t.Fatalf("provider '%v' is not supported: subscribe event-format is not defined", tt.provider)
			}
		})
	}
}
