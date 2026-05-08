package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

// Random ID with the PlatformEventChannelMember key prefix (0u3) — guaranteed
// not to match any real record in the org, so the GET inside deleteToSFAPI
// returns 404 and the DELETE is skipped.
const nonExistingMemberId = "0u3000000000000AAA"

// This script verifies that DeleteEventChannelMember is idempotent for
// non-existing records: passing an ID that doesn't exist should succeed (no
// error) because deleteToSFAPI now does an existence check first and treats a
// 404 as a no-op.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	resp, err := conn.DeleteEventChannelMember(ctx, nonExistingMemberId)
	if err != nil {
		logging.Logger(ctx).Error("DeleteEventChannelMember on non-existing record returned error", "error", err)
		return
	}

	fmt.Printf("DeleteEventChannelMember on non-existing record succeeded. resp=%v (expected nil)\n", resp)
}
