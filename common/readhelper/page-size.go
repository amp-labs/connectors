package readhelper

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// PageSizeWithDefaultStr returns the user-specified page size from params,
// converted to a string. If params.PageSize is empty value, defaultPageSize is returned instead.
func PageSizeWithDefaultStr(params common.ReadParams, defaultPageSize string) string {
	if params.PageSize <= 0 {
		return defaultPageSize
	}

	return strconv.Itoa(params.PageSize)
}
