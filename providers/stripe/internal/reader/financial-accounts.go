package reader

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// readTreasuryForMainAccount creates parallel read tasks for financial accounts
// belonging to the platform's main account.
//
// Each request includes the financial_account query parameter to identify the
// target Treasury financial account. Requests are executed in the context of
// the platform account and do not include a connected account header.
func (s *Strategy) readTreasuryForMainAccount(ctx context.Context,
	objectName string,
	fields datautils.StringSet,
	urlTemplate *urlbuilder.URL,
	financialAccounts map[string]FinancialAccount,
) (*common.ReadResult, error) {
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(financialAccounts))

	index := 0
	for _, financialAccount := range financialAccounts {
		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			taskUrl := urlTemplate.Clone()
			taskUrl.WithQueryParam("financial_account", financialAccount.Id)
			result, err := s.readRecords(ctx, objectName, fields, taskUrl)

			return financialAccount.Id, result, err
		}
		index += 1
	}

	return executeReadTasksTreasury(ctx, tasks, financialAccounts)
}

// readTreasuryForConnectedAccounts creates parallel read tasks for financial accounts
// belonging to connected accounts.
//
// Each request includes:
//   - the financial_account query parameter to identify the target Treasury account.
//   - the Stripe-Account header to execute the request in the context of the
//     connected account that owns the financial account.
func (s *Strategy) readTreasuryForConnectedAccounts(
	ctx context.Context,
	params common.ReadParams,
	urlTemplate *urlbuilder.URL,
	financialAccounts map[string]FinancialAccount,
) (*common.ReadResult, error) {
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(financialAccounts))

	index := 0
	for _, financialAccount := range financialAccounts {
		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			taskUrl := urlTemplate.Clone()
			taskUrl.WithQueryParam("financial_account", financialAccount.Id)
			header := makeConnectedAccountHeader(financialAccount.ConnectedAccountId)
			result, err := s.readRecords(ctx, params.ObjectName, params.Fields, taskUrl, header)

			return financialAccount.Id, result, err
		}
		index += 1
	}

	return executeReadTasksTreasury(ctx, tasks, financialAccounts)
}

// readNextPageTreasury resumes paginated reads for Treasury financial accounts.
//
// The aggregate token contains the next page URL and FinancialAccount context
// for each account that still has unread pages. This method uses that state to
// create independent requests for each financial account and executes them in
// parallel.
//
// The FinancialAccount context determines whether the request targets the
// platform account or a connected account. For connected accounts, the request
// includes the Stripe-Account header. For the platform account, no additional
// header is required.
//
// The financial_account query parameter is always added because of the
// Treasury API requirement.
//
// Example: a user may have paused reading the "treasury/inbound_transfers" object
// after processing some pages for 3 financial accounts owned by connected account
// "A" and 2 financial accounts owned by connected account "B". The aggregate token
// contains the next page position and account context for each of
// those 5 finnancial accounts.
//
// NOTE: Unlike readNextPageConnectedAccounts, tasks are indexed by Financial
// Account ID rather than Connected Account ID. Multiple financial accounts can
// belong to the same connected account, so Connected Account IDs are not unique
// task identifiers in this flow.
func (s *Strategy) readNextPageTreasury(ctx context.Context,
	params common.ReadParams,
	aggregateToken readhelper.AggregateNextPage[FinancialAccount],
) (*common.ReadResult, error) {
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(aggregateToken))
	financialAccounts := make(map[string]FinancialAccount)

	for index, token := range aggregateToken {
		financialAccount := token.Context
		financialAccounts[financialAccount.Id] = financialAccount
		nextPageToken := token.Value.String()

		urlTemplate, err := urlbuilder.New(nextPageToken)
		if err != nil {
			return nil, err
		}

		// Create parallel task for next page of each financial account.
		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			var headers []common.Header
			if financialAccount.ConnectedAccountId != "" {
				headers = append(headers, makeConnectedAccountHeader(financialAccount.ConnectedAccountId))
			}

			// financial_account query param should be already present.
			// Set it anyway to be explicit.
			taskUrl := urlTemplate.Clone()
			taskUrl.WithQueryParam("financial_account", financialAccount.Id)
			result, err := s.readRecords(ctx, params.ObjectName, params.Fields, taskUrl, headers...)

			return financialAccount.Id, result, err
		}
	}

	return executeReadTasksTreasury(ctx, tasks, financialAccounts)
}

// FinancialAccount represents an association between Connected Account and Financial account.
// When ConnectedAccountId is empty then this financial account belongs to the main platform account.
// Additionally, this struct is used to construct aggregate page token and that is the reason the `json` tags are used.
// This is a context for the next page token.
type FinancialAccount struct {
	Id                 string `json:"finAccId"`
	ConnectedAccountId string `json:"conAccId"`
}

type financialAccountsListResponse struct {
	HasMore bool `json:"has_more"`
	Data    []struct {
		Id string `json:"id"`
	} `json:"data"`
}

// listAllFinancialAccounts retrieves all financial accounts for the provided connected account IDs.
//
// If connectedAccountIds is nil or empty, the financial accounts for the platform account are returned.
func (s *Strategy) listAllFinancialAccounts(ctx context.Context,
	connectedAccountIds []string,
) (map[string]FinancialAccount, error) {
	// Retrieve financial accounts for the platform account.
	if len(connectedAccountIds) == 0 {
		return s.listFinancialAccountsForAccount(ctx, nil)
	}

	// Retrieve and aggregate financial accounts for each connected account.
	financialAccounts := make([]map[string]FinancialAccount, 0)

	for _, accountId := range connectedAccountIds {
		accounts, err := s.listFinancialAccountsForAccount(ctx, new(accountId))
		if err != nil {
			return nil, err
		}

		financialAccounts = append(financialAccounts, accounts)
	}

	return datautils.MergeMaps(financialAccounts...), nil
}

func (s *Strategy) listFinancialAccountsForAccount(ctx context.Context, // nolint:cyclop
	accountId *string,
) (map[string]FinancialAccount, error) {
	url, err := s.base.GetURL("treasury/financial_accounts")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	financialAccounts := make(map[string]FinancialAccount)

	for {
		headers := make([]common.Header, 0, 1)
		if accountId != nil {
			// Stripe requires the Stripe-Account header to make requests on behalf of a connected account.
			headers = append(headers, makeConnectedAccountHeader(*accountId))
		}

		res, err := s.base.JSONHTTPClient().Get(ctx, url.String(), headers...)
		if err != nil {
			return nil, err
		}

		accounts, err := common.UnmarshalJSON[financialAccountsListResponse](res)
		if err != nil {
			return nil, err
		}

		var lastItemId string

		for _, item := range accounts.Data {
			// Record the connected account that owns this financial account.
			var connectedAccountId string
			if accountId != nil {
				connectedAccountId = *accountId
			}

			lastItemId = item.Id
			financialAccounts[item.Id] = FinancialAccount{
				Id:                 item.Id,
				ConnectedAccountId: connectedAccountId,
			}
		}

		if !accounts.HasMore || len(financialAccounts) == 0 || lastItemId == "" {
			return financialAccounts, nil // => Desired return with collected ids.
		}

		// Prepare next page URL using the last account ID
		url.WithQueryParam("starting_after", lastItemId)
	}
}
