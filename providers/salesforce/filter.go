package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// RawFilterTypeSOQL is the raw filter type Salesforce supports:
// a SOQL condition, as it would appear after WHERE.
const RawFilterTypeSOQL = core.RawFilterTypeSOQL

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts raw filters of type RawFilterTypeSOQL with a string condition.
// Raw filters are only supported by the CRM module; Account Engagement (Pardot) rejects them.
func (c *Connector) ValidateRawFilter(filter common.RawFilter) error {
	if c.isPardotModule() {
		return fmt.Errorf("%w: not supported by module %s", common.ErrInvalidRawFilter, c.moduleID)
	}

	return core.ValidateRawFilter(filter)
}
