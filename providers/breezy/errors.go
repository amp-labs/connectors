package breezy

import "errors"

// ErrMissingCompanyID is returned when company_id metadata is required but not set.
var ErrMissingCompanyID = errors.New("breezy: company_id metadata is required")
