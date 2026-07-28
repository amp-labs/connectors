package deps

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// CDCOptimizationResolver resolves the Salesforce CDC quota-optimization opt-in configuration
// for a (project, group) pair.
type CDCOptimizationResolver interface {
	// GetCDCOptimizationConfig returns the CDC optimization config for the given project and
	// group, falling back to the project default when the group has no specific entry. Returns
	// nil when the project has no CDC optimization configured.
	GetCDCOptimizationConfig(ctx context.Context, projectID, groupRef string) *CDCOptimizationConfig
}

// CDCOptimizationConfig is the Salesforce CDC quota-optimization opt-in configuration for a
// group or a project. Mirrors the server's customer.CDCOptimizationConfig.
type CDCOptimizationConfig struct {
	// ManualCheckboxManagement reports whether the customer manually manages the checkbox field.
	ManualCheckboxManagement bool

	// ManualApexTriggerManagement reports whether the customer manually manages the Apex trigger.
	ManualApexTriggerManagement bool

	// ObjectEnabled maps object names to whether CDC optimization is enabled for them.
	ObjectEnabled map[common.ObjectName]bool
}
