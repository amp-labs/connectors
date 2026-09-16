package deps

import (
	"context"

	"github.com/amp-labs/amp-common/openapi"
	"github.com/amp-labs/connectors/common"
)

// SalesforceFlowResolver resolves the read fields each object's outbound message serializes, for
// record-triggered-flow subscribe.
type SalesforceFlowResolver interface {
	// GetSalesforceFlowSelectedFields returns the installation's resolved read fields per object,
	// as Salesforce API names.
	GetSalesforceFlowSelectedFields(
		ctx context.Context,
		inst *openapi.Installation,
		rev *openapi.Revision,
	) (map[common.ObjectName][]string, error)
}
