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

	// fieldConnectedAccountId is the field name used to store the connected account identifier
	// in ReadResult.Data[*].Fields.
	// This field is populated when ReadParamsOpts.ReadForAllConnectedAccounts is set to true.
	fieldConnectedAccountId = "AMPERSAND-connectedAccountId"

	// fieldFinancialAccountId is the field name used to store the financial account identifier
	// in ReadResult.Data[*].Fields.
	// This field is populated when read Object is part of the Treasury API.
	fieldFinancialAccountId = "AMPERSAND-financialAccountId"

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
