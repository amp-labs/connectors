package searchhelper

import (
	"fmt"
	"slices"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// ValidateRawFilterType returns the RawFilter type if it matches one of the allowed types.
// If RawFilter is nil, it returns an empty string.
func ValidateRawFilterType(params *common.SearchParams, allowedTypes ...string) (filterType string, err error) {
	if params.RawFilter == nil {
		return "", nil
	}

	if slices.Contains(allowedTypes, params.RawFilter.Type) {
		return params.RawFilter.Type, nil
	}

	return "", fmt.Errorf("%w: RawFilter.Type must be one of (%v)",
		common.ErrInvalidRawFilter,
		strings.Join(allowedTypes, ","),
	)
}
