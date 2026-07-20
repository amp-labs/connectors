package reader

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

// ReadScope defines the target scope for a read operation.
//
// It is the high-level branching point for the read execution flow.
// Non-Treasury scopes execute standard API reads against the platform account
// or connected accounts.
// Treasury scopes execute reads against Treasury financial accounts.
type ReadScope int

var ErrReadTargetUnknown = errors.New("unsupported read target")

const (
	ReadScopeMainAccount ReadScope = iota
	ReadScopeSelectedConnectedAccounts
	ReadScopeAllConnectedAccounts
	ReadScopeMainAccountTreasury
	ReadScopeSelectedConnectedAccountsTreasury
	ReadScopeAllConnectedAccountsTreasury
)

type readTarget struct {
	Scope      ReadScope
	AccountIDs []string
}

func inferReadTarget(params common.ReadParams) readTarget {
	opts, ok := params.Opts.(ReadParamsOpts)
	if !ok {
		if scopedObjectsForFinancialAccount.Has(params.ObjectName) {
			return readTarget{Scope: ReadScopeMainAccountTreasury}
		}

		return readTarget{Scope: ReadScopeMainAccount}
	}

	if len(opts.ReadForConnectedAccounts) > 0 {
		if scopedObjectsForFinancialAccount.Has(params.ObjectName) {
			return readTarget{Scope: ReadScopeSelectedConnectedAccountsTreasury}
		}

		return readTarget{
			Scope:      ReadScopeSelectedConnectedAccounts,
			AccountIDs: opts.ReadForConnectedAccounts,
		}
	}

	if opts.ReadForAllConnectedAccounts {
		if scopedObjectsForFinancialAccount.Has(params.ObjectName) {
			return readTarget{Scope: ReadScopeAllConnectedAccountsTreasury}
		}

		return readTarget{Scope: ReadScopeAllConnectedAccounts}
	}

	if scopedObjectsForFinancialAccount.Has(params.ObjectName) {
		return readTarget{Scope: ReadScopeMainAccountTreasury}
	}

	return readTarget{Scope: ReadScopeMainAccount}
}

// executeReadTasksAccounts executes multiple read tasks in parallel and aggregates their results.
//
// Used when reading objects across multiple Connected Accounts.
//
// Behavior:
//   - Each result row is annotated with its source Connected Account ID using fieldConnectedAccountId.
//   - If any task has remaining pages, an aggregate next page token is created
//     containing the next page state and Connected Account ID required to resume each unfinished read.
//   - Errors from all tasks are joined and returned as a single error.
func executeReadTasksAccounts(ctx context.Context,
	tasks []parallelfetch.Task[string, common.ReadResult],
) (*common.ReadResult, error) {
	result := parallelfetch.Execute(ctx, tasks, maxReadConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	return readhelper.AggregateReadResults(
		result.Records,
		func(accountId string) string {
			// Account ID uniquely identifies the source for next-page token resolution.
			// This will be used to construct Headers for the next page read operation.
			return accountId
		},
		func(accountId string, row *common.ReadResultRow) {
			// Enhance fields to indicate what account this row is associated with.
			row.Fields[fieldConnectedAccountId] = accountId
		},
	), nil
}

// executeReadTasksTreasury executes multiple read tasks in parallel and aggregates their results.
//
// Used when reading objects across Treasury financial accounts.
//
// Behavior:
//   - Each result row is annotated with its source Connected Account ID and
//     Financial Account ID using fieldConnectedAccountId and fieldFinancialAccountId.
//   - If any task has remaining pages, an aggregate next page token is created
//     containing the next page state and FinancialAccount context required to
//     resume each unfinished read. FinancialAccount context contains both the
//     Connected Account ID and Financial Account ID.
//   - Errors from all tasks are joined and returned as a single error.
//
// Input:
//   - tasks maps financial account IDs to the corresponding parallel requests.
//   - financialAccounts maps financial account IDs to their FinancialAccount
//     context used for pagination continuation.
func executeReadTasksTreasury(ctx context.Context,
	tasks []parallelfetch.Task[string, common.ReadResult],
	financialAccounts map[string]FinancialAccount,
) (*common.ReadResult, error) {
	result := parallelfetch.Execute(ctx, tasks, maxReadConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	return readhelper.AggregateReadResults(
		result.Records,
		func(financialAccountId string) FinancialAccount {
			// Financial Account ID uniquely identifies the source for next-page token resolution.
			// This will be used to construct Headers/Query Params for the next page read operation.
			return financialAccounts[financialAccountId]
		},
		func(financialAccountId string, row *common.ReadResultRow) {
			// Enhance fields to indicate what financial and connected accounts this row is associated with.
			row.Fields[fieldConnectedAccountId] = financialAccounts[financialAccountId].ConnectedAccountId
			row.Fields[fieldFinancialAccountId] = financialAccountId
		},
	), nil
}
