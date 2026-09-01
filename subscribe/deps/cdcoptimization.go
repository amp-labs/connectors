package deps

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
)

// CDCOptimizationResolver resolves the Salesforce CDC quota-optimization opt-in for one
// installation.
//
// The opt-in is expressed per subscribe object in the installation config, so answering takes
// the caller's config-over-revision resolution — which is why this is a seam rather than
// something this package derives from the installation itself.
type CDCOptimizationResolver interface {
	// GetCDCOptimizationConfig returns the resolved CDC optimization config for an installation.
	//
	// Nil means the caller has no opinion: leave whatever is provisioned at the provider alone.
	// A non-nil config is the authoritative desired state, so one listing no objects is an
	// instruction to tear existing quota optimization down — not the same answer as nil.
	//
	// An error means resolution failed and the desired state is unknown. Callers must not read
	// that as "nothing enabled", or a transient failure would deprovision a live optimization.
	GetCDCOptimizationConfig(
		ctx context.Context,
		inst *openapi.Installation,
		rev *openapi.Revision,
	) (*CDCOptimizationConfig, error)
}

// CDCOptimizationConfig is the resolved Salesforce CDC quota-optimization desired state for one
// installation.
type CDCOptimizationConfig struct {
	// ManualCheckboxManagement reports whether the customer manually manages the checkbox field.
	ManualCheckboxManagement bool

	// ManualApexTriggerManagement reports whether the customer manually manages the Apex trigger.
	ManualApexTriggerManagement bool

	// EnabledObjects lists the objects opted into CDC quota optimization. Only enabled objects
	// are represented: once the caller has resolved the config, "explicitly disabled" and "never
	// mentioned" are the same desired state, and an object's absence here is what drives teardown.
	EnabledObjects []common.ObjectName
}
