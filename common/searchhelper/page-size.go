package searchhelper

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
)

// PageSizeWithDefaultStr returns the user-specified page size from params,
// converted to a string. If params.PageSize is empty value, defaultPageSize is returned instead.
func PageSizeWithDefaultStr(params *common.SearchParams, defaultPageSize string) string {
	if params.Limit <= 0 {
		return defaultPageSize
	}

	return strconv.FormatInt(params.Limit, 10)
}

// PageSizeWithDefault returns the user-specified page size from params.
// If params.PageSize is empty value, defaultPageSize is returned instead.
func PageSizeWithDefault(params *common.SearchParams, defaultPageSize int64) int64 {
	if params.Limit <= 0 {
		return defaultPageSize
	}

	return params.Limit
}
