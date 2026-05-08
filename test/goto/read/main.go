package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	connTest "github.com/amp-labs/connectors/test/goto"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetGoToConnector(ctx, providers.ModuleGoTo)

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "webinars",
		Fields:     connectors.Fields("accountKey", "omid"),
	})
	if err != nil {
		utils.Fail("error reading from GoTo", "error", err)
	}

	fmt.Println("Reading webinars..")
	utils.DumpJSON(res, os.Stdout)

	res, err = conn.Read(ctx, common.ReadParams{
		ObjectName: "historicalMeetings",
		Fields:     connectors.Fields("accountKey", "meetingId"),
	})
	if err != nil {
		utils.Fail("error reading from GoTo", "error", err)
	}

	fmt.Println("Reading historical meetings..")
	utils.DumpJSON(res, os.Stdout)

}
