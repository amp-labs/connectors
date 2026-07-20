package reader

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// https://docs.stripe.com/api/connected-accounts
func makeConnectedAccountHeader(accountID string) common.Header {
	return common.Header{
		Key:   "Stripe-Account",
		Value: accountID,
	}
}

func (s *Strategy) readForConnectedAccounts(
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
			result, err := s.readRecords(ctx, params.ObjectName, params.Fields, url, header)

			return accountID, result, err
		}
	}

	return executeReadTasksAccounts(ctx, tasks)
}

// readNextPageConnectedAccounts resumes paginated reads for connected accounts.
// It uses an aggregate token to track the next page position for each account
// and parallelizes the requests.
func (s *Strategy) readNextPageConnectedAccounts(ctx context.Context,
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
			result, err := s.readRecords(ctx, params.ObjectName, params.Fields, url, header)

			return accountID, result, err
		}
	}

	return executeReadTasksAccounts(ctx, tasks)
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
func (s *Strategy) listAllAccounts(ctx context.Context) ([]string, error) {
	url, err := s.base.GetURL("accounts")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))

	accountIDs := make([]string, 0)

	for {
		res, err := s.base.JSONHTTPClient().Get(ctx, url.String())
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
