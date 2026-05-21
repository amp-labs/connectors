package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/tools/debug"
)

// Hardcoded FullName used to provoke a DUPLICATE_DEVELOPER_NAME on the second
// CreateEventChannel call. Stable across runs so the script is repeatable —
// the trailing cleanup deletes the channel after each run. The "__chn" suffix
// is required for custom PlatformEventChannel records; without it Salesforce
// treats the request as targeting a standard channel and rejects Create.
const duplicateChannelFullName = "amp_duplicate_create_test__chn"

// This script verifies that CreateEventChannel is idempotent on
// DUPLICATE_DEVELOPER_NAME: calling Create twice with the same FullName
// should succeed both times and return the same record (the second call
// recovers via SOQL+GET instead of erroring out).
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetSalesforceConnector(ctx)
	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	channel := &salesforce.EventChannel{
		FullName: duplicateChannelFullName,
		Metadata: &salesforce.EventChannelMetadata{
			ChannelType: "data",
			Label:       duplicateChannelFullName,
		},
	}

	first, err := conn.CreateEventChannel(ctx, channel)
	if err != nil {
		logging.Logger(ctx).Error("first CreateEventChannel failed", "error", err)
		return
	}

	fmt.Printf("First CreateEventChannel succeeded. id=%s\n", first.Id)

	// Second call with the same FullName should hit DUPLICATE_DEVELOPER_NAME
	// and recover by looking up the existing record.
	second, err := conn.CreateEventChannel(ctx, &salesforce.EventChannel{
		FullName: duplicateChannelFullName,
		Metadata: &salesforce.EventChannelMetadata{
			ChannelType: "data",
			Label:       duplicateChannelFullName,
		},
	})
	if err != nil {
		logging.Logger(ctx).Error("second CreateEventChannel (duplicate) failed", "error", err)
		return
	}

	fmt.Printf("Second CreateEventChannel (duplicate) recovered. id=%s\n", second.Id)

	if first.Id != second.Id {
		logging.Logger(ctx).Error("recovered record id mismatch",
			"firstId", first.Id, "secondId", second.Id)
		return
	}

	fmt.Println("First: ", debug.PrettyFormatStringJSON(first))
	fmt.Println("Second: ", debug.PrettyFormatStringJSON(second))

	fmt.Println("IDs match — duplicate recovery returned the existing record.")

	// Clean up so the script is repeatable.
	if _, err := conn.DeleteEventChannel(ctx, first.Id); err != nil {
		logging.Logger(ctx).Error("cleanup DeleteEventChannel failed", "error", err)
		return
	}

	fmt.Println("Cleanup successful")
}
