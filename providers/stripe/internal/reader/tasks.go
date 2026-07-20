package reader

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/parallelfetch"
)

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
