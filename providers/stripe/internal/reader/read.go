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

// Read retrieves a list of records for a given object type.
//
// Supported features:
//   - NextPage: Supports pagination for objects that Stripe paginates.
//   - AssociatedObjects: Fetches nested objects by specifying fields to expand.
//     See [Stripe expanding objects docs](https://docs.stripe.com/api/expanding_objects).
//   - Incremental Reading: Not supported (the `Since` parameter is ignored).
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
	aggregateToken, ok := readhelper.GetAggregateToken[string](params.NextPage)
	if !ok {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return s.readRecords(ctx, params.ObjectName, params.Fields, url)
	}

	// (2) connected accounts (resume parallelized reads)
	return s.readNextPageConnectedAccounts(ctx, params, aggregateToken)
}

// readFirstPage reads the first page of object records for either the main account
// or all connected accounts (if ReadForAllConnectedAccounts is enabled).
func (s *Strategy) readFirstPage(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
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
		accountIDs, err := s.listAllAccounts(ctx)
		if err != nil {
			return nil, err
		}

		return s.readForConnectedAccounts(ctx, params, url, accountIDs)
	default:
		return nil, fmt.Errorf("%w: %v", ErrReadTargetUnknown, target.Scope)
	}
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
