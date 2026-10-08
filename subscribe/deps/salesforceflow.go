package deps

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
)

// SalesforceFlowResolver resolves the read fields each object's outbound message serializes, for
// record-triggered-flow subscribe.
type SalesforceFlowResolver interface {
	// GetSalesforceFlowConfig returns the installation's resolved flow-subscribe configuration.
	GetSalesforceFlowConfig(
		ctx context.Context,
		inst *openapi.Installation,
		rev *openapi.Revision,
	) (SalesforceFlowConfig, error)
}

// SalesforceFlowConfig is the resolved flow-subscribe configuration for one installation.
type SalesforceFlowConfig struct {
	// SelectedFields maps each subscribed object to its read fields, as Salesforce API names.
	SelectedFields map[common.ObjectName][]string
}
