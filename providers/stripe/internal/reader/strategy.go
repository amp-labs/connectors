package reader

import (
	"github.com/amp-labs/connectors/providers/stripe/internal/core"
)

const (
	// maxReadConcurrency limits concurrent requests to avoid exceeding Stripe's rate limit of 100 requests/second.
	// Set to 3 as a safe conservative value.
	//
	// Rate limit: [https://docs.stripe.com/rate-limits](https://docs.stripe.com/rate-limits)
	maxReadConcurrency = 3

	// fieldConnectedAccountID is the field name used to store the connected account identifier
	// in ReadResult.Data[*].Fields.
	// This field is populated when ReadParamsOpts.ReadForAllConnectedAccounts is set to true.
	fieldConnectedAccountID = "AMPERSAND-connectedAccountId"

	// DefaultPageSize is number of elements per page.
	DefaultPageSize = 100
)

type Strategy struct {
	base *core.Base
}

func NewStrategy(base *core.Base) *Strategy {
	return &Strategy{
		base: base,
	}
}
