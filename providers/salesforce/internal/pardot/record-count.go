package pardot

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
)

// ErrRecordCountNotImplemented is returned when GetRecordCount is called on the Pardot adapter.
// Record count for the Account Engagement (Pardot) module is not yet implemented.
var ErrRecordCountNotImplemented = errors.New("record count is not implemented for Pardot (Account Engagement) module")

// GetRecordCount returns an error because record count for Pardot is not yet implemented.
func (a *Adapter) GetRecordCount(
	_ context.Context,
	_ *common.RecordCountParams,
) (*common.RecordCountResult, error) {
	return nil, ErrRecordCountNotImplemented
}
