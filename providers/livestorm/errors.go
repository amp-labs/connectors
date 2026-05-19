package livestorm

import "errors"

var (
	// ErrJobIDRequired is returned when reading jobs without a job id in ReadParams.Filter.
	ErrJobIDRequired = errors.New("read jobs requires a non-empty ReadParams.Filter (job id)")
)
