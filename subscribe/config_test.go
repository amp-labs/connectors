package subscribe

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/subscribe/deps"
)

func boolPtr(v bool) *bool {
	return &v
}

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
				Registration:   boolPtr(true),
				SubscribeByAPI: boolPtr(true),
			},
		},
	}

	topLevelReqs := &providers.SubscribeRequirements{
		Registration:   boolPtr(false),
		SubscribeByAPI: boolPtr(false),
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
			SubscribeRequirements: &providers.SubscribeRequirements{SubscribeByAPI: boolPtr(true)},
		},
	}

	topLevelReqs := &providers.SubscribeRequirements{SubscribeByAPI: boolPtr(false)}
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

	topLevelReqs := &providers.SubscribeRequirements{Registration: boolPtr(true)}
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

	topLevelReqs := &providers.SubscribeRequirements{Registration: boolPtr(true)}
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
