package reader

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/stripe/internal/metadata"
)

// ReadParamsOpts defines optional parameters for the Read operation.
type ReadParamsOpts struct {
	// ReadForConnectedAccounts enables reading data for specified connected accounts.
	// This takes precedence over ReadForAllConnectedAccounts.
	ReadForConnectedAccounts []string
	// ReadForAllConnectedAccounts enables reading data from all connected accounts
	// instead of only the main account. When true, the connector parallelizes reads
	// across all connected accounts and adds the connected account ID to each result row.
	ReadForAllConnectedAccounts bool
}

// Read retrieves a list of records for the specified object name.
//
// Supported features:
//   - NextPage: Pagination is supported. For Connected Accounts and Treasury
//     financial accounts, next-page tokens use aggregate contextual format.
//   - Fields: JSONPath-based field selection and on-demand expansion are
//     supported. See https://docs.stripe.com/api/expanding_objects.
//   - Incremental Reading: Not supported; the "Since" parameter is ignored.
//
// Domain:
//   - Objects owned by the main platform account and Connected Accounts are supported.
//   - Treasury API objects are read through their associated financial accounts.
//
// Treasury reads:
//   - Treasury objects require a financial_account ID in the request.
//   - The connector first calls:
//     https://docs.stripe.com/api/treasury/financial_accounts/list
//   - It then calls the object endpoint for each financial account, for example:
//     https://docs.stripe.com/api/treasury/transactions/list
//     https://docs.stripe.com/api/treasury/inbound_transfers/list
//     https://docs.stripe.com/api/treasury/outbound_transfers/list
//     https://docs.stripe.com/api/treasury/outbound_payments/list
//     https://docs.stripe.com/api/treasury/credit_reversals/list
//     https://docs.stripe.com/api/treasury/debit_reversals/list
//   - For Connected Accounts, the same flow is used with the Stripe-Account
//     header set to the connected account ID.
//
// Rate limits and concurrency:
//   - Stripe rate limits apply per API key and per connected account:
//     https://docs.stripe.com/rate-limits
//   - Financial accounts are discovered sequentially.
//   - Treasury reads are executed in parallel up to maxReadConcurrency.
//   - Lower maxReadConcurrency if rate-limit errors occur.
//
// Parallelization:
//   - Connected accounts and financial accounts are discovered sequentially.
//   - Requests are then parallelized to improve throughput.
//
// Relevant Stripe documentation:
//   - Expanding objects fields: https://docs.stripe.com/api/expanding_objects
//   - Connected accounts: https://docs.stripe.com/api/connected-accounts
//   - Financial Accounts List: https://docs.stripe.com/api/treasury/financial_accounts/list
//   - Treasury overview: https://docs.stripe.com/treasury
func (s *Strategy) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	// Handle first-page reading for either:
	// (1) main account, or
	// (2) all connected accounts (if ReadForAllConnectedAccounts is true)
	if params.IsFirstPage() {
		return s.readFirstPage(ctx, params)
	}

	// Handle next-page reading for either:
	// (1) main account (regular pagination)
	// (2) connected accounts (resume parallelized reads)
	return s.readNextPage(ctx, params)
}

// readFirstPage reads the first page of object records for the requested scope.
//
// The read target determines whether the request is executed against the main
// platform account, selected connected accounts, all connected accounts, or
// Treasury financial accounts. Treasury reads have a separate execution flow
// because requests are scoped by financial account and may require connected
// account context.
func (s *Strategy) readFirstPage(ctx context.Context, // nolint:cyclop
	params common.ReadParams,
) (*common.ReadResult, error) {
	url, err := s.buildFirstPageReadURL(params)
	if err != nil {
		return nil, err
	}

	target := inferReadTarget(params)
	switch target.Scope {
	case ReadScopeMainAccount:
		// Standard read for the current account.
		return s.readRecords(ctx, params.ObjectName, params.Fields, url)
	case ReadScopeSelectedConnectedAccounts:
		return s.readForConnectedAccounts(ctx, params, url, target.AccountIDs)
	case ReadScopeAllConnectedAccounts:
		accountIds, err := s.listAllAccounts(ctx)
		if err != nil {
			return nil, err
		}

		return s.readForConnectedAccounts(ctx, params, url, accountIds)
	case ReadScopeMainAccountTreasury:
		financialAccounts, err := s.listAllFinancialAccounts(ctx, nil)
		if err != nil {
			return nil, err
		}

		return s.readTreasuryForMainAccount(ctx, params.ObjectName, params.Fields, url, financialAccounts)
	case ReadScopeSelectedConnectedAccountsTreasury:
		financialAccounts, err := s.listAllFinancialAccounts(ctx, target.AccountIDs)
		if err != nil {
			return nil, err
		}

		return s.readTreasuryForConnectedAccounts(ctx, params, url, financialAccounts)
	case ReadScopeAllConnectedAccountsTreasury:
		accountIds, err := s.listAllAccounts(ctx)
		if err != nil {
			return nil, err
		}

		financialAccounts, err := s.listAllFinancialAccounts(ctx, accountIds)
		if err != nil {
			return nil, err
		}

		return s.readTreasuryForConnectedAccounts(ctx, params, url, financialAccounts)
	default:
		return nil, fmt.Errorf("%w: %v", ErrReadTargetUnknown, target.Scope)
	}
}

// readNextPage resumes a paginated read using the provided next page token.
//
// The pagination flow depends on the type of token:
//   - Aggregate tokens with string context resume reads across Connected Accounts.
//   - Aggregate tokens with FinancialAccount context resume reads across Treasury
//     financial accounts.
//   - Standard next page URLs resume regular reads for objects owned by the main
//     platform account.
//
// Aggregate pagination tokens contain the context required to reconstruct the
// request for each account when multiple account scopes are being read in
// parallel.
func (s *Strategy) readNextPage(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if token, ok := readhelper.GetAggregateToken[string](params.NextPage); ok {
		return s.readNextPageConnectedAccounts(ctx, params, token)
	}

	if token, ok := readhelper.GetAggregateToken[FinancialAccount](params.NextPage); ok {
		return s.readNextPageTreasury(ctx, params, token)
	}

	// Default read of the next page.
	// It is neither for connected account nor is it for Treasury API.
	url, err := urlbuilder.New(params.NextPage.String())
	if err != nil {
		return nil, err
	}

	return s.readRecords(ctx, params.ObjectName, params.Fields, url)
}

// readRecords performs a GET request to the specified URL and parses the response.
// Optional headers can be provided (e.g., for connected account authentication).
func (s *Strategy) readRecords(ctx context.Context,
	objectName string,
	selectedFields datautils.StringSet,
	url *urlbuilder.URL,
	headers ...common.Header,
) (*common.ReadResult, error) {
	res, err := s.base.JSONHTTPClient().Get(ctx, url.String(), headers...)
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(common.ModuleRoot, objectName)

	return common.ParseResult(res,
		makeGetRecords(responseFieldName),
		makeNextRecordsURL(url),
		readhelper.MakeMarshaledSelectedDataFunc(
			fieldsSelector,
			jsonquery.Convertor.ObjectToMap,
		),
		selectedFields,
	)
}
