package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/auphan"
	"github.com/amp-labs/connectors/test/utils"
)

// Known close_date totals on 2026-11-05 for the store in auphan-creds.json: payment_type -> {payment, tips, count}.
// A missing tips value means Auphan returned null.
var expected = map[string][3]string{ //nolint:gochecknoglobals
	"Visa":          {"4680.61", "437.37", "153"},
	"MC":            {"1411.96", "180.76", "36"},
	"Doordash":      {"509.46", "0.00", "20"},
	"Cash":          {"490.46", "", "21"},
	"Amex":          {"232.70", "25.77", "6"},
	"Online Visa":   {"156.56", "9.83", "7"},
	"Discover Card": {"144.48", "16.52", "4"},
	"Online MC":     {"92.16", "4.33", "2"},
	"UberEats":      {"89.32", "0.00", "4"},
	"GrubHub":       {"16.10", "0.00", "1"},
}

func main() {
	os.Exit(run())
}

func run() int {
	ctx := context.Background()
	conn := auphan.GetAuphanConnector(ctx)

	// Times are sent as written, in their own location: the store timezone is unknown.
	day := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "payments",
		Fields:     connectors.Fields("date", "payment_type", "gl_code", "payment", "tips", "count"),
		Since:      day,
		Until:      day.Add(24*time.Hour - time.Second),
	})
	if err != nil {
		utils.Fail("error reading payments", "error", err)
	}

	out, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(out))

	return compare(res.Data)
}

func compare(rows []common.ReadResultRow) int {
	failed := 0
	seen := map[string]bool{}
	totals := [3]*big.Rat{new(big.Rat), new(big.Rat), new(big.Rat)}

	for _, row := range rows {
		paymentType, _ := row.Fields["payment_type"].(string)
		got := [3]string{text(row.Fields["payment"]), text(row.Fields["tips"]), text(row.Fields["count"])}

		want, ok := expected[paymentType]
		if !ok {
			fmt.Printf("UNEXPECTED row %q %v\n", paymentType, got)

			failed++

			continue
		}

		seen[paymentType] = true

		if got != want {
			fmt.Printf("MISMATCH %q got %v want %v\n", paymentType, got, want)

			failed++
		}

		for i, value := range got {
			if value != "" {
				amount, _ := new(big.Rat).SetString(value)
				totals[i].Add(totals[i], amount)
			}
		}
	}

	for paymentType := range expected {
		if !seen[paymentType] {
			fmt.Printf("MISSING %q\n", paymentType)

			failed++
		}
	}

	fmt.Printf("totals: payment=%s tips=%s count=%s (want 7823.81 674.58 254)\n",
		totals[0].FloatString(2), totals[1].FloatString(2), totals[2].FloatString(0))

	if totals[0].FloatString(2) != "7823.81" || totals[1].FloatString(2) != "674.58" ||
		totals[2].FloatString(0) != "254" {
		failed++
	}

	if failed > 0 {
		fmt.Printf("FAIL: %d problems\n", failed)

		return 1
	}

	fmt.Println("OK: all 10 payment types match the known totals")

	return 0
}

func text(value any) string {
	if value == nil {
		return ""
	}

	return fmt.Sprint(value)
}
