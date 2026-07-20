package reader

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestReadForTreasury(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseAccountsFirstPage := testutils.DataFromFile(t, "accounts/1-first-page.json")
	responseAccountsLastPage := testutils.DataFromFile(t, "accounts/2-last-page.json")
	responseAcc1FinAccounts1 := testutils.DataFromFile(t, "financial-account/con-acc-1/1-first-page.json")
	responseAcc1FinAccounts2 := testutils.DataFromFile(t, "financial-account/con-acc-1/2-last-page.json")
	responseAcc2FinAccounts1 := testutils.DataFromFile(t, "financial-account/con-acc-2/1-last-page.json")
	responseAcc1Fin00aTransactions := testutils.DataFromFile(t, "transaction_entries/con-acc-1-fin-a.json")
	responseAcc1Fin00bTransactions := testutils.DataFromFile(t, "transaction_entries/con-acc-1-fin-b.json")
	responseAcc2Fin00cTransactions := testutils.DataFromFile(t, "transaction_entries/con-acc-2-fin-c.json")

	tests := []testconn.TestCaseRead{
		{
			Name: "Read Treasury transactions for all connected accounts",
			// Read all connected accounts (2 pages).
			// Read all financial accounts (2 pages).
			// Account 1 has Fin00a and Fin00b.
			// Account 2 has Fin00c.
			// Read transactions for Acc1 Fin00a (1 page).
			// Read transactions for Acc1 Fin00b (1 page).
			// Read transactions for Acc2 Fin00c (1 page).
			Input: common.ReadParams{
				ObjectName: "treasury/transaction_entries",
				Fields:     connectors.Fields("id"),
				Opts:       ReadParamsOpts{ReadForAllConnectedAccounts: true},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{
					// ========
					// Connected accounts
					// ========
					{
						If: mockcond.And{
							mockcond.Path("/v1/accounts"),
							mockcond.QueryParamsMissing("starting_after"),
						},
						Then: mockserver.Response(http.StatusOK, responseAccountsFirstPage),
					},
					{
						If: mockcond.And{
							mockcond.Path("/v1/accounts"),
							mockcond.QueryParam("starting_after", "acct_1Nv0FGQ9RKHgCVdK"),
						},
						Then: mockserver.Response(http.StatusOK, responseAccountsLastPage),
					},
					// ========
					// Financial accounts (for Connected Account 1)
					// ========
					{
						If: mockcond.And{
							mockcond.Path("/v1/treasury/financial_accounts"),
							mockcond.QueryParamsMissing("starting_after"), // first page
							mockcond.Header(map[string][]string{
								"Stripe-Account": {"acct_1Nv0FGQ9RKHgCVdK"}, // Account #1
							}),
						},
						Then: mockserver.Response(http.StatusOK, responseAcc1FinAccounts1),
					},
					{
						If: mockcond.And{
							mockcond.Path("/v1/treasury/financial_accounts"),
							mockcond.QueryParam("starting_after", "finAcc_00a_a3738395bf7b"),
							mockcond.Header(map[string][]string{
								"Stripe-Account": {"acct_1Nv0FGQ9RKHgCVdK"}, // Account #1
							}),
						},
						Then: mockserver.Response(http.StatusOK, responseAcc1FinAccounts2),
					},
					// ========
					// Financial accounts (for Connected Account 1)
					// ========
					{
						If: mockcond.And{
							mockcond.Path("/v1/treasury/financial_accounts"),
							mockcond.QueryParamsMissing("starting_after"),
							mockcond.Header(map[string][]string{
								"Stripe-Account": {"acct_2c81631b20648a90"}, // Account #2
							}),
						},
						Then: mockserver.Response(http.StatusOK, responseAcc2FinAccounts1),
					},
					// ========
					// Transactions for Connected Account 1
					// ========
					{
						If: mockcond.And{
							mockcond.Path("/v1/treasury/transaction_entries"),
							mockcond.QueryParam("financial_account", "finAcc_00a_a3738395bf7b"), // Fin #A
							mockcond.Header(map[string][]string{
								"Stripe-Account": {"acct_1Nv0FGQ9RKHgCVdK"}, // Account #1
							}),
						},
						Then: mockserver.Response(http.StatusOK, responseAcc1Fin00aTransactions),
					},
					{
						If: mockcond.And{
							mockcond.Path("/v1/treasury/transaction_entries"),
							mockcond.QueryParam("financial_account", "finAcc_00b_a9981a9bdebe"), // Fin #B
							mockcond.Header(map[string][]string{
								"Stripe-Account": {"acct_1Nv0FGQ9RKHgCVdK"}, // Account #1
							}),
						},
						Then: mockserver.Response(http.StatusOK, responseAcc1Fin00bTransactions),
					},
					// ========
					// Transactions for Connected Account 2
					// ========
					{
						If: mockcond.And{
							mockcond.Path("/v1/treasury/transaction_entries"),
							mockcond.QueryParam("financial_account", "finAcc_00c_a9981a9bdebe"), // Fin #C
							mockcond.Header(map[string][]string{
								"Stripe-Account": {"acct_2c81631b20648a90"}, // Account #2
							}),
						},
						Then: mockserver.Response(http.StatusOK, responseAcc2Fin00cTransactions),
					},
				},
			}.Server(),
			Comparator: testconn.ComparatorSubsetReadSorted,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"AMPERSAND-connectedAccountId": "acct_1Nv0FGQ9RKHgCVdK",
						"id":                           "transaction_0a0_b6317b07df58",
					},
					Raw: map[string]any{"object": "treasury.transaction_entry"},
					Id:  "transaction_0a0_b6317b07df58",
				}, {
					Fields: map[string]any{
						"AMPERSAND-connectedAccountId": "acct_1Nv0FGQ9RKHgCVdK",
						"id":                           "transaction_0b0_76f1bb782252",
					},
					Raw: map[string]any{"object": "treasury.transaction_entry"},
					Id:  "transaction_0b0_76f1bb782252",
				}, {
					Fields: map[string]any{
						"AMPERSAND-connectedAccountId": "acct_2c81631b20648a90",
						"id":                           "transaction_0c0_ea539fa34e3e",
					},
					Raw: map[string]any{"object": "treasury.transaction_entry"},
					Id:  "transaction_0c0_ea539fa34e3e",
				}},
				// We are not done reading
				NextPage: `[{
					"context": {
						"conAccId": "acct_1Nv0FGQ9RKHgCVdK",
						"finAccId": "finAcc_00a_a3738395bf7b"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00a_a3738395bf7b&limit=100&starting_after=transaction_0a0_b6317b07df58"
				}, {
					"context": {
						"conAccId": "acct_1Nv0FGQ9RKHgCVdK",
						"finAccId": "finAcc_00b_a9981a9bdebe"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00b_a9981a9bdebe&limit=100&starting_after=transaction_0b0_76f1bb782252"
				}, {
					"context": {
						"conAccId": "acct_2c81631b20648a90",
						"finAccId": "finAcc_00c_a9981a9bdebe"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00c_a9981a9bdebe&limit=100&starting_after=transaction_0c0_ea539fa34e3e"
				}]`,
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read next page Treasury transactions with aggregate token",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id"),
				Opts:       ReadParamsOpts{ReadForAllConnectedAccounts: true},
				NextPage: `[{
					"context": {
						"conAccId": "acct_1Nv0FGQ9RKHgCVdK",
						"finAccId": "finAcc_00a_a3738395bf7b"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00a_a3738395bf7b&limit=100&starting_after=transaction_0a0_b6317b07df58"
				}, {
					"context": {
						"conAccId": "acct_1Nv0FGQ9RKHgCVdK",
						"finAccId": "finAcc_00b_a9981a9bdebe"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00b_a9981a9bdebe&limit=100&starting_after=transaction_0b0_76f1bb782252"
				}, {
					"context": {
						"conAccId": "acct_2c81631b20648a90",
						"finAccId": "finAcc_00c_a9981a9bdebe"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00c_a9981a9bdebe&limit=100&starting_after=transaction_0c0_ea539fa34e3e"
				}]`,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/v1/treasury/transaction_entries"),
						mockcond.QueryParam("financial_account", "finAcc_00a_a3738395bf7b"),
						mockcond.QueryParam("starting_after", "transaction_0a0_b6317b07df58"),
						mockcond.Header(map[string][]string{"Stripe-Account": {"acct_1Nv0FGQ9RKHgCVdK"}}),
					},
					Then: mockserver.ResponseString(http.StatusOK, `{"data":[{"id":"1"}]}`),
				}, {
					If: mockcond.And{
						mockcond.Path("/v1/treasury/transaction_entries"),
						mockcond.QueryParam("financial_account", "finAcc_00b_a9981a9bdebe"),
						mockcond.QueryParam("starting_after", "transaction_0b0_76f1bb782252"),
						mockcond.Header(map[string][]string{"Stripe-Account": {"acct_1Nv0FGQ9RKHgCVdK"}}),
					},
					Then: mockserver.ResponseString(http.StatusOK, `{"data":[{"id":"2"}], "has_more":true}`),
				}, {
					If: mockcond.And{
						mockcond.Path("/v1/treasury/transaction_entries"),
						mockcond.QueryParam("financial_account", "finAcc_00c_a9981a9bdebe"),
						mockcond.QueryParam("starting_after", "transaction_0c0_ea539fa34e3e"),
						mockcond.Header(map[string][]string{"Stripe-Account": {"acct_2c81631b20648a90"}}),
					},
					Then: mockserver.ResponseString(http.StatusOK, `{"data":[{"id":"3"}]}`),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetReadSorted,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"AMPERSAND-connectedAccountId": "acct_1Nv0FGQ9RKHgCVdK",
						"id":                           "1",
					},
					Raw: map[string]any{"id": "1"}, Id: "1",
				}, {
					Fields: map[string]any{
						"AMPERSAND-connectedAccountId": "acct_1Nv0FGQ9RKHgCVdK",
						"id":                           "2",
					},
					Raw: map[string]any{"id": "2"}, Id: "2",
				}, {
					Fields: map[string]any{
						"AMPERSAND-connectedAccountId": "acct_2c81631b20648a90",
						"id":                           "3",
					},
					Raw: map[string]any{"id": "3"}, Id: "3",
				}},
				NextPage: `[{
					"context": {
						"conAccId": "acct_1Nv0FGQ9RKHgCVdK",
						"finAccId": "finAcc_00b_a9981a9bdebe"
					},
					"value": "` + testconn.URLTestServer + `/v1/treasury/transaction_entries?financial_account=finAcc_00b_a9981a9bdebe&limit=100&starting_after=2"
				}]`,
				Done: false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}
