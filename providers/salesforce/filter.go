package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

var _ connectors.RawFilterConnector = (*Connector)(nil)

// ValidateRawFilter accepts Salesforce's raw filter syntax, core.SOQLFilter (type "soql").
// The syntax itself lives in core so the CRM search strategy can use it too; this method only adds
// what depends on the connector: raw filters are supported by the CRM module, and Account Engagement
// (Pardot) rejects them.
func (c *Connector) ValidateRawFilter(filter common.RawFilter) error {
	if c.isPardotModule() {
		return fmt.Errorf("%w: not supported by module %s", common.ErrInvalidRawFilter, c.moduleID)
	}

	_, err := common.RawFilterAs[core.SOQLFilter](filter)

	return err
}
