package stripe

import (
	"context"
	"errors"
	"maps"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/stripe/metadata"
	"github.com/spyzhov/ajson"
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
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	// Handle first-page reading for either:
	// (1) main account, or
	// (2) all connected accounts (if ReadForAllConnectedAccounts is true)
	if params.IsFirstPage() {
		return c.readFirstPage(ctx, params)
	}

	// Handle next-page reading for either:
	// (1) main account (regular pagination)
	aggregateToken, ok := readhelper.GetAggregateToken[string](params.NextPage)
	if !ok {
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		return c.readRecords(ctx, params.ObjectName, params.Fields, url)
	}

	// (2) connected accounts (resume parallelized reads)
	return c.readNextPageConnectedAccounts(ctx, params, aggregateToken)
}

// readFirstPage reads the first page of object records for either the main account
// or all connected accounts (if ReadForAllConnectedAccounts is enabled).
func (c *Connector) readFirstPage(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	url, err := c.buildFirstPageReadURL(params)
	if err != nil {
		return nil, err
	}

	target := inferReadTarget(params)
	switch target.Scope {
	case ReadScopeMainAccount:
		// Standard read for the current account.
		return c.readRecords(ctx, params.ObjectName, params.Fields, url)
	case ReadScopeSelectedConnectedAccounts:
		return c.readForConnectedAccounts(ctx, params, url, target.AccountIDs)
	case ReadScopeAllConnectedAccounts:
		accountIDs, err := c.listAllAccounts(ctx)
		if err != nil {
			return nil, err
		}

		return c.readForConnectedAccounts(ctx, params, url, accountIDs)
	default:
		return nil, fmt.Errorf("%w: %v", ErrReadTargetUnknown, target.Scope)
	}
}

func (c *Connector) readForConnectedAccounts(
	ctx context.Context,
	params common.ReadParams,
	url *urlbuilder.URL,
	accountIDs []string,
) (*common.ReadResult, error) {
	// Parallelized read across all connected accounts.
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(accountIDs))
	for index, accountID := range accountIDs {
		// Each task reads records for a specific connected account
		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			header := makeConnectedAccountHeader(accountID)
			result, err := c.readRecords(ctx, params.ObjectName, params.Fields, url, header)

			return accountID, result, err
		}
	}

	return executeReadTasks(ctx, tasks)
}

// buildFirstPageReadURL constructs the API request URL for a Read operation.
// It applies:
//   - Pagination: adds "limit" query parameter
//   - Incremental reading: adds "created[gte]" for objects supporting incremental reads
//   - Object expansion: adds "expand[]" for fields of format:
//     => "$['line_items']['currency']"
//     => "$['line_items']['description']"
//     => "$['line_items']['...']"
//     => "$['source']['payment_intent']['customer']['id']"
//     => "$['source']['payment_intent']['id']"
//
// See [Stripe expand documentation](https://docs.stripe.com/expand#how-it-works).
func (c *Connector) buildFirstPageReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, strconv.Itoa(DefaultPageSize))
	url.WithQueryParam("limit", pageSize)

	if !params.Since.IsZero() && incrementalObjects.Has(params.ObjectName) {
		url.WithQueryParam("created[gte]", strconv.FormatInt(params.Since.Unix(), 10))
	}

	expandTargets := make(datautils.Set[string])

	for field := range params.Fields {
		// If the requested field supports expansion, add it to the expand list
		// so the provider can return the nested object in the response.
		queryParam := metadata.MakeExpandableQueryParam(params.ObjectName, field)
		if queryParam != "" {
			expandTargets.AddOne(queryParam)
		}
	}

	if len(expandTargets) != 0 {
		// Expand nested objects by adding "data.<field>" to expand[] query parameter
		url.WithQueryParamList("expand[]", expandTargets.List())
	}

	return url, nil
}

// executeReadTasks runs multiple read tasks in parallel and aggregates their results.
// Used when reading for multiple connected accounts.
//
// Behavior:
//   - Each result row is marked with its source connected account ID via fieldConnectedAccountID
//   - Returns an aggregated next page token if any individual read has more data
//   - Joins all errors and returns them as a single error
func executeReadTasks(ctx context.Context,
	tasks []parallelfetch.Task[string, common.ReadResult],
) (*common.ReadResult, error) {
	result := parallelfetch.Execute(ctx, tasks, maxReadConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	return readhelper.AggregateReadResults(
		result.Records,
		func(accountID string) string {
			// Account ID uniquely identifies the source for next-page token resolution.
			// This will be used to construct Headers for the next page read operation.
			return accountID
		},
		func(accountID string, row *common.ReadResultRow) {
			// Enhance fields to indicate what account this row is associated with.
			row.Fields[fieldConnectedAccountID] = accountID
		},
	), nil
}

// readNextPageConnectedAccounts resumes paginated reads for connected accounts.
// It uses an aggregate token to track the next page position for each account
// and parallelizes the requests.
func (c *Connector) readNextPageConnectedAccounts(ctx context.Context,
	params common.ReadParams,
	aggregateToken readhelper.AggregateNextPage[string],
) (*common.ReadResult, error) {
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(aggregateToken))
	for index, token := range aggregateToken {
		accountID := token.Context
		nextPageToken := token.Value.String()

		url, err := urlbuilder.New(nextPageToken)
		if err != nil {
			return nil, err
		}

		// Create parallel task for this account's next page
		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			header := makeConnectedAccountHeader(accountID)
			result, err := c.readRecords(ctx, params.ObjectName, params.Fields, url, header)

			return accountID, result, err
		}
	}

	return executeReadTasks(ctx, tasks)
}

// readRecords performs a GET request to the specified URL and parses the response.
// Optional headers can be provided (e.g., for connected account authentication).
func (c *Connector) readRecords(ctx context.Context,
	objectName string,
	selectedFields datautils.StringSet,
	url *urlbuilder.URL,
	headers ...common.Header,
) (*common.ReadResult, error) {
	res, err := c.Client.Get(ctx, url.String(), headers...)
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectName)

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

// makeGetRecords creates a NodeRecordsFunc that extracts records from Stripe's API response.
// It retrieves the array field containing the list of records (e.g., "data" for most objects).
func makeGetRecords(responseFieldName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

type accountsListResponse struct {
	HasMore bool `json:"has_more"`
	Data    []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// listAllAccounts retrieves all connected account IDs for the main account.
// It handles pagination internally to collect all accounts.
//
// See [Stripe accounts list documentation](https://docs.stripe.com/api/accounts/list).
func (c *Connector) listAllAccounts(ctx context.Context) ([]string, error) {
	url, err := c.getURL("accounts")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	accountIDs := make([]string, 0)

	for {
		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		accounts, err := common.UnmarshalJSON[accountsListResponse](res)
		if err != nil {
			return nil, err
		}

		for _, item := range accounts.Data {
			accountIDs = append(accountIDs, item.ID)
		}

		if !accounts.HasMore || len(accountIDs) == 0 {
			return accountIDs, nil // => Desired return with collected ids.
		}

		// Prepare next page URL using the last account ID
		lastItemID := accountIDs[len(accountIDs)-1]
		url.WithQueryParam("starting_after", lastItemID)
	}
}

type ReadScope int

var ErrReadTargetUnknown = errors.New("unsupported read target")

const (
	ReadScopeMainAccount ReadScope = iota
	ReadScopeSelectedConnectedAccounts
	ReadScopeAllConnectedAccounts
)

type readTarget struct {
	Scope      ReadScope
	AccountIDs []string
}

func inferReadTarget(params common.ReadParams) readTarget {
	opts, ok := params.Opts.(ReadParamsOpts)
	if !ok {
		return readTarget{Scope: ReadScopeMainAccount}
	}

	if len(opts.ReadForConnectedAccounts) > 0 {
		return readTarget{
			Scope:      ReadScopeSelectedConnectedAccounts,
			AccountIDs: opts.ReadForConnectedAccounts,
		}
	}

	if opts.ReadForAllConnectedAccounts {
		return readTarget{Scope: ReadScopeAllConnectedAccounts}
	}

	return readTarget{Scope: ReadScopeMainAccount}
}

// incrementalObjects contains object names that support incremental reading
// via the "created[gte]" query parameter.
var incrementalObjects = datautils.NewSet( // nolint:gochecknoglobals
	"accounts",
	"application_fees",
	"balance/history",
	"balance_transactions",
	"charges",
	"checkout/sessions",
	"coupons",
	"credit_notes",
	"customers",
	"disputes",
	"events",
	"file_links",
	"files",
	"forwarding/requests",
	"identity/verification_reports",
	"identity/verification_sessions",
	"invoiceitems",
	"invoices",
	"issuing/authorizations",
	"issuing/cardholders",
	"issuing/cards",
	"issuing/disputes",
	"issuing/transactions",
	"payment_intents",
	"payouts",
	"plans",
	"prices",
	"products",
	"promotion_codes",
	"radar/early_fraud_warnings",
	"radar/value_lists",
	"refunds",
	"reporting/report_runs",
	"reviews",
	"setup_intents",
	"shipping_rates",
	"subscription_schedules",
	"subscriptions",
	"tax_rates",
	"topups",
	"transfers",
	"treasury/financial_accounts",
)

func fieldsSelector(node *ajson.Node, fields []string) (map[string]any, string, error) {
	root, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, "", err
	}

	identifier, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, "", err
	}

	customFields, err := getCustomFields(node)
	if err != nil {
		return nil, "", err
	}

	selected := readhelper.SelectFields(root, datautils.NewSetFromList(fields))
	maps.Copy(selected, customFields)

	return selected, identifier, nil
}
