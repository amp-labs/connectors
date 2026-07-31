package subscribe

import (
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

var errMaintenanceUnsupportedProvider = errors.New("maintenance is not supported for this provider")

// maintenancePeriods maps providers to how often their webhook subscriptions need
// renewal. For example, Gmail watch subscriptions expire after 7 days, so maintenance
// runs daily to renew them before expiry.
//
// This is the renewal *interval*, which has no ProviderInfo equivalent today; whether maintenance
// is required at all is derived from ProviderInfo (see ShouldPerformMaintenance).
//
//nolint:mnd
var maintenancePeriods = map[providers.Provider]time.Duration{
	providers.Zoho:   time.Hour * 24 * 7,
	providers.Google: time.Hour * 24,
}

// GetMaintenancePeriod returns the scheduled maintenance interval for a provider's webhooks.
// Returns an error if the provider does not require periodic maintenance.
func GetMaintenancePeriod(provider providers.Provider) (time.Duration, error) {
	if period, ok := maintenancePeriods[provider]; ok {
		return period, nil
	}

	return 0, fmt.Errorf("%w: %s", errMaintenanceUnsupportedProvider, provider)
}

// MaintenanceConfig is the subcomponent of ProviderConfig that exposes the periodic-maintenance
// queries for a provider's webhook subscriptions: whether maintenance is required at all, and the
// renewal interval when it is.
//
// The "should perform maintenance?" answer is derived from ProviderInfo.SubscribeRequirements.
// Maintenance, so it is computed by ShouldPerform rather than stored. The renewal interval, by
// contrast, has no ProviderInfo equivalent today, so it is declared per-provider via the
// renewalInterval field — mirroring how RegistrationConfig keeps the non-derivable emptyResultFn /
// buildParamsFn declarative while computing IsRequired.
//
// module / providerInfo are bound at GetProviderConfig time and threaded into ShouldPerform. The
// zero MaintenanceConfig is valid; ProviderConfig embeds it by value, and the methods themselves
// take a value receiver, so callers can invoke them directly without nil-checking — the type
// system guarantees a non-nil receiver. A zero renewalInterval means the provider does not
// declare periodic maintenance.
type MaintenanceConfig struct {
	// renewalInterval is how often the provider's webhook subscriptions must be renewed (e.g.
	// Gmail watch subscriptions expire after 7 days). Declared per-provider because there is no
	// ProviderInfo equivalent today. Zero means the provider declares no maintenance interval.
	renewalInterval time.Duration

	// module and providerInfo are bound by GetProviderConfig at call time. ShouldPerform reads
	// them to compute the answer from ProviderInfo.
	module       common.ModuleID
	providerInfo *providers.ProviderInfo
}

// ShouldPerform reports whether the provider's webhook subscriptions require periodic maintenance
// (renewal). Delegates to ShouldPerformMaintenance.
func (m MaintenanceConfig) ShouldPerform() bool {
	return ShouldPerformMaintenance(m.module, m.providerInfo)
}

// Interval returns the renewal interval declared for the provider's webhooks and true, or
// (0, false) when the provider declares no periodic maintenance. Idiomatic use:
//
//	if interval, ok := cfg.Maintenance.Interval(); ok { ... }
func (m MaintenanceConfig) Interval() (time.Duration, bool) {
	if m.renewalInterval <= 0 {
		return 0, false
	}

	return m.renewalInterval, true
}
