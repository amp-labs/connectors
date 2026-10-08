package reader

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestInferReadTarget(t *testing.T) {
	t.Parallel()

	const (
		regularObject  = "customers"
		treasuryObject = "treasury/transactions"
	)

	tests := []struct {
		name   string
		params common.ReadParams
		want   readTarget
	}{
		{
			name: "Main account when options are absent",
			params: common.ReadParams{
				ObjectName: regularObject,
			},
			want: readTarget{
				Scope: ReadScopeMainAccount,
			},
		},
		{
			name: "Main Treasury account when options are absent",
			params: common.ReadParams{
				ObjectName: treasuryObject,
			},
			want: readTarget{
				Scope: ReadScopeMainAccountTreasury,
			},
		},
		{
			name: "Main account when options have their zero value",
			params: common.ReadParams{
				ObjectName: regularObject,
				Opts:       ReadParamsOpts{},
			},
			want: readTarget{
				Scope: ReadScopeMainAccount,
			},
		},
		{
			name: "Main Treasury account when options have their zero value",
			params: common.ReadParams{
				ObjectName: treasuryObject,
				Opts:       ReadParamsOpts{},
			},
			want: readTarget{
				Scope: ReadScopeMainAccountTreasury,
			},
		},
		{
			name: "Selected connected accounts for a regular object",
			params: common.ReadParams{
				ObjectName: regularObject,
				Opts: ReadParamsOpts{
					ReadForConnectedAccounts: []string{
						"acct_connected_1",
						"acct_connected_2",
					},
				},
			},
			want: readTarget{
				Scope: ReadScopeSelectedConnectedAccounts,
				AccountIDs: []string{
					"acct_connected_1",
					"acct_connected_2",
				},
			},
		},
		{
			name: "Selected connected accounts for a Treasury object",
			params: common.ReadParams{
				ObjectName: treasuryObject,
				Opts: ReadParamsOpts{
					ReadForConnectedAccounts: []string{
						"acct_connected_1",
						"acct_connected_2",
					},
				},
			},
			want: readTarget{
				Scope: ReadScopeSelectedConnectedAccountsTreasury,
				AccountIDs: []string{
					"acct_connected_1",
					"acct_connected_2",
				},
			},
		},
		{
			name: "All connected accounts for a regular object",
			params: common.ReadParams{
				ObjectName: regularObject,
				Opts: ReadParamsOpts{
					ReadForAllConnectedAccounts: true,
				},
			},
			want: readTarget{
				Scope: ReadScopeAllConnectedAccounts,
			},
		},
		{
			name: "All connected accounts for a Treasury object",
			params: common.ReadParams{
				ObjectName: treasuryObject,
				Opts: ReadParamsOpts{
					ReadForAllConnectedAccounts: true,
				},
			},
			want: readTarget{
				Scope: ReadScopeAllConnectedAccountsTreasury,
			},
		},
		{
			name: "Selected connected accounts take precedence over all connected accounts",
			params: common.ReadParams{
				ObjectName: regularObject,
				Opts: ReadParamsOpts{
					ReadForConnectedAccounts: []string{
						"acct_connected_1",
					},
					ReadForAllConnectedAccounts: true,
				},
			},
			want: readTarget{
				Scope: ReadScopeSelectedConnectedAccounts,
				AccountIDs: []string{
					"acct_connected_1",
				},
			},
		},
		{
			name: "Selected Treasury connected accounts take precedence over all connected accounts",
			params: common.ReadParams{
				ObjectName: treasuryObject,
				Opts: ReadParamsOpts{
					ReadForConnectedAccounts: []string{
						"acct_connected_1",
					},
					ReadForAllConnectedAccounts: true,
				},
			},
			want: readTarget{
				Scope: ReadScopeSelectedConnectedAccountsTreasury,
				AccountIDs: []string{
					"acct_connected_1",
				},
			},
		},
		{
			name: "Empty selected account list does not select connected accounts",
			params: common.ReadParams{
				ObjectName: regularObject,
				Opts: ReadParamsOpts{
					ReadForConnectedAccounts: []string{},
				},
			},
			want: readTarget{
				Scope: ReadScopeMainAccount,
			},
		},
		{
			name: "Empty selected account list falls back to all connected accounts",
			params: common.ReadParams{
				ObjectName: regularObject,
				Opts: ReadParamsOpts{
					ReadForConnectedAccounts:    []string{},
					ReadForAllConnectedAccounts: true,
				},
			},
			want: readTarget{
				Scope: ReadScopeAllConnectedAccounts,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := inferReadTarget(tt.params)

			result := testutils.NewCompareResult()
			result.Assert("readTarget", tt.want, got)
			result.Validate(t, tt.name)
		})
	}
}
